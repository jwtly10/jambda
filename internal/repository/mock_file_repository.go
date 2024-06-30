package repository

import (
	"github.com/jwtly10/jambda/api/data"
	"github.com/stretchr/testify/mock"
)

// MockFunctionRepository is a mock implementation of FileRepository
type MockFunctionRepository struct {
	mock.Mock
}

func (m *MockFunctionRepository) InitFunctionEntity(externalId string) (*data.FunctionEntity, error) {
	args := m.Called(externalId)
	return nil, args.Error(0)
}

func (m *MockFunctionRepository) GetFunctionEntityFromExternalId(externalId string) (*data.FunctionEntity, error) {
	args := m.Called(externalId)
	return nil, args.Error(0)
}

func (m *MockFunctionRepository) GetConfigurationFromExternalId(externalId string) (*data.FunctionConfig, error) {
	args := m.Called(externalId)
	return nil, args.Error(0)
}
