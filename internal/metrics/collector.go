package metrics

import "sync"

type MetricCollector struct {
	requestCounts map[string]int
	mutex         sync.Mutex
}

func (mc *MetricCollector) IncrementRequestCount(functionId string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.requestCounts[functionId]++
}

func (mc *MetricCollector) GetRequestCount(functionId string) int {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	return mc.requestCounts[functionId]
}

func (mc *MetricCollector) ResetRequestCount(functionId string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.requestCounts[functionId] = 0

}
