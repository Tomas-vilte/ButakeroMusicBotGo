package file_storage

import (
	"github.com/stretchr/testify/mock"
)

type MockStatePersistent struct {
	mock.Mock
}

func (m *MockStatePersistent) ReadState(filepath string) (*FileState, error) {
	args := m.Called(filepath)
	return args.Get(0).(*FileState), args.Error(1)
}

func (m *MockStatePersistent) WriteState(filepath string, state *FileState) error {
	args := m.Called(filepath, state)
	return args.Error(0)
}
