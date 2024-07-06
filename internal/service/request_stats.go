package service

import (
	"sync"
	"time"

	"github.com/jwtly10/jambda/internal/logging"
)

type FunctionStats struct {
	RequestCount int
	LastRequest  time.Time
}

type RequestStatsService struct {
	functionStats map[string]*FunctionStats
	mu            sync.Mutex
	log           logging.Logger
	ds            DockerService
	fs            FunctionService
}

func NewRequestStatsService(log logging.Logger, ds DockerService, fs FunctionService) *RequestStatsService {
	return &RequestStatsService{
		functionStats: make(map[string]*FunctionStats),
		log:           log,
		ds:            ds,
		fs:            fs,
	}
}

func (rs *RequestStatsService) IncrementRequestCount(functionID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.functionStats[functionID]; !exists {
		rs.functionStats[functionID] = &FunctionStats{}
	}
	rs.functionStats[functionID].RequestCount++
	rs.functionStats[functionID].LastRequest = time.Now()
}

func (rs *RequestStatsService) ResetRequestCount(functionID string, skipLock bool) {
	// This function is only called from within a function that already has a lock acquired
	// So technically this is not needed.
	// However, if in future i look to reset some stats about a function
	// I want to ensure I dont accidently try to acquire a lock during another lock.
	// Will revise this at some point
	if !skipLock {
		rs.mu.Lock()
		defer rs.mu.Unlock()
	}
	if stats, exists := rs.functionStats[functionID]; exists {
		stats.RequestCount = 0
	}
}

func (rs *RequestStatsService) GetFunctionStats() map[string]*FunctionStats {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	// Return a copy of the stats to avoid exposing internal state
	copyStats := make(map[string]*FunctionStats)
	for id, stats := range rs.functionStats {
		copyStats[id] = &FunctionStats{
			RequestCount: stats.RequestCount,
			LastRequest:  stats.LastRequest,
		}
	}
	return copyStats
}

func (rs *RequestStatsService) ScaleDownUnusedContainers(idleDuration time.Duration) {
	rs.log.Info("Checking for idle functions")
	functionIDs, _ := rs.fs.GetAllActiveFunctions()
	rs.mu.Lock()
	defer rs.mu.Unlock()
	currentTime := time.Now()
	for _, function := range functionIDs {
		rs.log.Infof("Checking function ID  %s", function.ExternalId)
		if stats, exists := rs.functionStats[function.ExternalId]; exists {
			rs.log.Infof("Stats for function '%s', LastRequest: '%s', ReqCount: '%d'", function.ExternalId, stats.LastRequest.String(), stats.RequestCount)
			// We only turn off functions if they are beyond the idle duration
			if currentTime.Sub(stats.LastRequest) > idleDuration && stats.RequestCount > 0 {
				rs.log.Infof("Found idle function '%s', LastRequest: '%s', ReqCount: '%d'", function.ExternalId, stats.LastRequest, stats.RequestCount)
				rs.ds.StopContainerForFunction(function.ExternalId)
				rs.ResetRequestCount(function.ExternalId, true)
			}
		}
	}
}
