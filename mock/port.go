// Code generated by MockGen. DO NOT EDIT.
// Source: internal/core/port/port.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
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

// MockResultRepository is a mock of ResultRepository interface.
type MockResultRepository struct {
	ctrl     *gomock.Controller
	recorder *MockResultRepositoryMockRecorder
}

// MockResultRepositoryMockRecorder is the mock recorder for MockResultRepository.
type MockResultRepositoryMockRecorder struct {
	mock *MockResultRepository
}

// NewMockResultRepository creates a new mock instance.
func NewMockResultRepository(ctrl *gomock.Controller) *MockResultRepository {
	mock := &MockResultRepository{ctrl: ctrl}
	mock.recorder = &MockResultRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResultRepository) EXPECT() *MockResultRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockResultRepository) Create(result *domain.JobResult) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", result)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockResultRepositoryMockRecorder) Create(result interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockResultRepository)(nil).Create), result)
}

// Delete mocks base method.
func (m *MockResultRepository) Delete(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockResultRepositoryMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockResultRepository)(nil).Delete), id)
}

// Get mocks base method.
func (m *MockResultRepository) Get(id string) (*domain.JobResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", id)
	ret0, _ := ret[0].(*domain.JobResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockResultRepositoryMockRecorder) Get(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockResultRepository)(nil).Get), id)
}

// MockJobQueue is a mock of JobQueue interface.
type MockJobQueue struct {
	ctrl     *gomock.Controller
	recorder *MockJobQueueMockRecorder
}

// MockJobQueueMockRecorder is the mock recorder for MockJobQueue.
type MockJobQueueMockRecorder struct {
	mock *MockJobQueue
}

// NewMockJobQueue creates a new mock instance.
func NewMockJobQueue(ctrl *gomock.Controller) *MockJobQueue {
	mock := &MockJobQueue{ctrl: ctrl}
	mock.recorder = &MockJobQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJobQueue) EXPECT() *MockJobQueueMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockJobQueue) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockJobQueueMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockJobQueue)(nil).Close))
}

// Pop mocks base method.
func (m *MockJobQueue) Pop() <-chan *domain.Job {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pop")
	ret0, _ := ret[0].(<-chan *domain.Job)
	return ret0
}

// Pop indicates an expected call of Pop.
func (mr *MockJobQueueMockRecorder) Pop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pop", reflect.TypeOf((*MockJobQueue)(nil).Pop))
}

// Push mocks base method.
func (m *MockJobQueue) Push(j *domain.Job) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", j)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Push indicates an expected call of Push.
func (mr *MockJobQueueMockRecorder) Push(j interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockJobQueue)(nil).Push), j)
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
func (m *MockJobService) Create(name, taskName, description string, timeout int, metadata interface{}) (*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", name, taskName, description, timeout, metadata)
	ret0, _ := ret[0].(*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockJobServiceMockRecorder) Create(name, taskName, description, timeout, metadata interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockJobService)(nil).Create), name, taskName, description, timeout, metadata)
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

// Update mocks base method.
func (m *MockJobService) Update(id, name, description string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", id, name, description)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockJobServiceMockRecorder) Update(id, name, description interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockJobService)(nil).Update), id, name, description)
}

// MockResultService is a mock of ResultService interface.
type MockResultService struct {
	ctrl     *gomock.Controller
	recorder *MockResultServiceMockRecorder
}

// MockResultServiceMockRecorder is the mock recorder for MockResultService.
type MockResultServiceMockRecorder struct {
	mock *MockResultService
}

// NewMockResultService creates a new mock instance.
func NewMockResultService(ctrl *gomock.Controller) *MockResultService {
	mock := &MockResultService{ctrl: ctrl}
	mock.recorder = &MockResultServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResultService) EXPECT() *MockResultServiceMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockResultService) Delete(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockResultServiceMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockResultService)(nil).Delete), id)
}

// Get mocks base method.
func (m *MockResultService) Get(id string) (*domain.JobResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", id)
	ret0, _ := ret[0].(*domain.JobResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockResultServiceMockRecorder) Get(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockResultService)(nil).Get), id)
}

// MockWorkService is a mock of WorkService interface.
type MockWorkService struct {
	ctrl     *gomock.Controller
	recorder *MockWorkServiceMockRecorder
}

// MockWorkServiceMockRecorder is the mock recorder for MockWorkService.
type MockWorkServiceMockRecorder struct {
	mock *MockWorkService
}

// NewMockWorkService creates a new mock instance.
func NewMockWorkService(ctrl *gomock.Controller) *MockWorkService {
	mock := &MockWorkService{ctrl: ctrl}
	mock.recorder = &MockWorkServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkService) EXPECT() *MockWorkServiceMockRecorder {
	return m.recorder
}

// CreateWork mocks base method.
func (m *MockWorkService) CreateWork(j *domain.Job) domain.Work {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWork", j)
	ret0, _ := ret[0].(domain.Work)
	return ret0
}

// CreateWork indicates an expected call of CreateWork.
func (mr *MockWorkServiceMockRecorder) CreateWork(j interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWork", reflect.TypeOf((*MockWorkService)(nil).CreateWork), j)
}

// Exec mocks base method.
func (m *MockWorkService) Exec(ctx context.Context, w domain.Work) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exec", ctx, w)
	ret0, _ := ret[0].(error)
	return ret0
}

// Exec indicates an expected call of Exec.
func (mr *MockWorkServiceMockRecorder) Exec(ctx, w interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockWorkService)(nil).Exec), ctx, w)
}

// Send mocks base method.
func (m *MockWorkService) Send(w domain.Work) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Send", w)
}

// Send indicates an expected call of Send.
func (mr *MockWorkServiceMockRecorder) Send(w interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockWorkService)(nil).Send), w)
}

// Start mocks base method.
func (m *MockWorkService) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockWorkServiceMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockWorkService)(nil).Start))
}

// Stop mocks base method.
func (m *MockWorkService) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockWorkServiceMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockWorkService)(nil).Stop))
}

// MockConsumerService is a mock of ConsumerService interface.
type MockConsumerService struct {
	ctrl     *gomock.Controller
	recorder *MockConsumerServiceMockRecorder
}

// MockConsumerServiceMockRecorder is the mock recorder for MockConsumerService.
type MockConsumerServiceMockRecorder struct {
	mock *MockConsumerService
}

// NewMockConsumerService creates a new mock instance.
func NewMockConsumerService(ctrl *gomock.Controller) *MockConsumerService {
	mock := &MockConsumerService{ctrl: ctrl}
	mock.recorder = &MockConsumerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConsumerService) EXPECT() *MockConsumerServiceMockRecorder {
	return m.recorder
}

// Consume mocks base method.
func (m *MockConsumerService) Consume() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Consume")
}

// Consume indicates an expected call of Consume.
func (mr *MockConsumerServiceMockRecorder) Consume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Consume", reflect.TypeOf((*MockConsumerService)(nil).Consume))
}

// Stop mocks base method.
func (m *MockConsumerService) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockConsumerServiceMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockConsumerService)(nil).Stop))
}
