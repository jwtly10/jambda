package repository

import (
	"github.com/jwtly10/jambda/api/data"
	"github.com/stretchr/testify/mock"
)

// MockFileRepository is a mock implementation of FileRepository
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) InitFileMetaData(externalId string) (*data.FileEntity, error) {
	args := m.Called(externalId)
	return nil, args.Error(0)
}

func (m *MockFileRepository) GetFileFromExternalId(externalId string) (*data.FileEntity, error) {
	args := m.Called(externalId)
	return nil, args.Error(0)
}
