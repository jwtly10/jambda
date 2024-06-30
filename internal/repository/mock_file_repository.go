package repository

import (
	"github.com/stretchr/testify/mock"
)

// MockFileRepository is a mock implementation of FileRepository
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) InitFileMetaData(externalId string) error {
	args := m.Called(externalId)
	return args.Error(0)
}

func (m *MockFileRepository) GetFileFromExternalId(externalId string) (*FileEntity, error) {
	args := m.Called(externalId)
	return nil, args.Error(0)
}
