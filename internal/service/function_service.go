package service

import (
	"fmt"
	"net/http"

	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/errors"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/repository"
)

type FunctionService struct {
	repo repository.IFunctionRepository
	log  logging.Logger
	fs   FileService
	cv   ConfigValidator
}

func NewFunctionService(repo repository.IFunctionRepository, log logging.Logger, fs FileService, cv ConfigValidator) *FunctionService {
	return &FunctionService{
		log:  log,
		repo: repo,
		fs:   fs,
		cv:   cv,
	}
}

// UploadFunction uploads a new function by processing the binary and saving the file and configuration for function.
func (fs *FunctionService) UploadFunction(r *http.Request) (*data.FunctionEntity, error) {
	return fs.fs.ProcessNewFunction(r)
}

func (fs *FunctionService) UpdateConfig(externalId, name string, config *data.FunctionConfig) (*data.FunctionEntity, error) {
	fs.log.Infof("Updating config for function '%s'", externalId)

	// Validate the new config
	err := fs.cv.ValidateConfig(config)
	if err != nil {
		fs.log.Errorf("Config validation failed %v", err)
		return nil, errors.NewValidationError(fmt.Sprintf("error validating config json: %v", err))
	}

	res, err := fs.repo.UpdateConfigByExternalId(externalId, name, *config)
	if err != nil {
		fs.log.Error("Failed to retrieve function: ", err)
		return nil, errors.NewInternalError(fmt.Sprintf("error updating new function config to db: %v", err))
	}

	return res, nil
}

func (fs *FunctionService) GetAllActiveFunctions() ([]data.FunctionEntity, error) {
	fs.log.Info("Getting all active functions")
	functions, err := fs.repo.GetAllActiveFunctions()
	if err != nil {
		fs.log.Error("Failed to retrieve functions: ", err)
		return nil, errors.NewInternalError(fmt.Sprintf("error retrieving active functions from db: %v", err))
	}

	if functions == nil {
		return []data.FunctionEntity{}, nil
	}

	return functions, nil
}

func (fs *FunctionService) DeleteFunction(externalId string) error {
	fs.log.Infof("Deleting function '%s'", externalId)
	err := fs.repo.DeleteFunctionByExternalId(externalId)
	if err != nil {
		fs.log.Error("Failed to delete function: ", err)
		return errors.NewInternalError(fmt.Sprintf("error deleting function from db: %v", err))
	}
	// Now should also delete from the file system too TODO
	return nil
}
