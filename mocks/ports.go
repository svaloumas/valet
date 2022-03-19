// Code generated by MockGen. DO NOT EDIT.
// Source: internal/core/ports/ports.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	domain "valet/internal/core/domain"

	gomock "github.com/golang/mock/gomock"
)

// MockJobRepository is a mock of JobRepository interface.
type MockJobRepository struct {
	ctrl     *gomock.Controller
	recorder *MockJobRepositoryMockRecorder
}

// MockJobRepositoryMockRecorder is the mock recorder for MockJobRepository.
type MockJobRepositoryMockRecorder struct {
	mock *MockJobRepository
}

// NewMockJobRepository creates a new mock instance.
func NewMockJobRepository(ctrl *gomock.Controller) *MockJobRepository {
	mock := &MockJobRepository{ctrl: ctrl}
	mock.recorder = &MockJobRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJobRepository) EXPECT() *MockJobRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockJobRepository) Create(j *domain.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", j)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockJobRepositoryMockRecorder) Create(j interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockJobRepository)(nil).Create), j)
}

// Delete mocks base method.
func (m *MockJobRepository) Delete(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockJobRepositoryMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockJobRepository)(nil).Delete), id)
}

// Get mocks base method.
func (m *MockJobRepository) Get(id string) (*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", id)
	ret0, _ := ret[0].(*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockJobRepositoryMockRecorder) Get(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockJobRepository)(nil).Get), id)
}

// Update mocks base method.
func (m *MockJobRepository) Update(id string, j *domain.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", id, j)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockJobRepositoryMockRecorder) Update(id, j interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockJobRepository)(nil).Update), id, j)
}

// MockJobService is a mock of JobService interface.
type MockJobService struct {
	ctrl     *gomock.Controller
	recorder *MockJobServiceMockRecorder
}

// MockJobServiceMockRecorder is the mock recorder for MockJobService.
type MockJobServiceMockRecorder struct {
	mock *MockJobService
}

// NewMockJobService creates a new mock instance.
func NewMockJobService(ctrl *gomock.Controller) *MockJobService {
	mock := &MockJobService{ctrl: ctrl}
	mock.recorder = &MockJobServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJobService) EXPECT() *MockJobServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockJobService) Create(name, description string) (*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", name, description)
	ret0, _ := ret[0].(*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockJobServiceMockRecorder) Create(name, description interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockJobService)(nil).Create), name, description)
}

// Delete mocks base method.
func (m *MockJobService) Delete(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockJobServiceMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockJobService)(nil).Delete), id)
}

// Get mocks base method.
func (m *MockJobService) Get(id string) (*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", id)
	ret0, _ := ret[0].(*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockJobServiceMockRecorder) Get(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockJobService)(nil).Get), id)
}
