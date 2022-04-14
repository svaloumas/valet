// Code generated by MockGen. DO NOT EDIT.
// Source: internal/core/port/port.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"
	time "time"
	domain "valet/internal/core/domain"
	work "valet/internal/core/service/worksrv/work"

	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// CheckHealth mocks base method.
func (m *MockStorage) CheckHealth() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckHealth")
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckHealth indicates an expected call of CheckHealth.
func (mr *MockStorageMockRecorder) CheckHealth() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckHealth", reflect.TypeOf((*MockStorage)(nil).CheckHealth))
}

// Close mocks base method.
func (m *MockStorage) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStorageMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStorage)(nil).Close))
}

// CreateJob mocks base method.
func (m *MockStorage) CreateJob(j *domain.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateJob", j)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateJob indicates an expected call of CreateJob.
func (mr *MockStorageMockRecorder) CreateJob(j interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateJob", reflect.TypeOf((*MockStorage)(nil).CreateJob), j)
}

// CreateJobResult mocks base method.
func (m *MockStorage) CreateJobResult(result *domain.JobResult) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateJobResult", result)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateJobResult indicates an expected call of CreateJobResult.
func (mr *MockStorageMockRecorder) CreateJobResult(result interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateJobResult", reflect.TypeOf((*MockStorage)(nil).CreateJobResult), result)
}

// DeleteJob mocks base method.
func (m *MockStorage) DeleteJob(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJob", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJob indicates an expected call of DeleteJob.
func (mr *MockStorageMockRecorder) DeleteJob(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJob", reflect.TypeOf((*MockStorage)(nil).DeleteJob), id)
}

// DeleteJobResult mocks base method.
func (m *MockStorage) DeleteJobResult(jobID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJobResult", jobID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJobResult indicates an expected call of DeleteJobResult.
func (mr *MockStorageMockRecorder) DeleteJobResult(jobID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJobResult", reflect.TypeOf((*MockStorage)(nil).DeleteJobResult), jobID)
}

// GetDueJobs mocks base method.
func (m *MockStorage) GetDueJobs() ([]*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDueJobs")
	ret0, _ := ret[0].([]*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDueJobs indicates an expected call of GetDueJobs.
func (mr *MockStorageMockRecorder) GetDueJobs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDueJobs", reflect.TypeOf((*MockStorage)(nil).GetDueJobs))
}

// GetJob mocks base method.
func (m *MockStorage) GetJob(id string) (*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJob", id)
	ret0, _ := ret[0].(*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJob indicates an expected call of GetJob.
func (mr *MockStorageMockRecorder) GetJob(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJob", reflect.TypeOf((*MockStorage)(nil).GetJob), id)
}

// GetJobResult mocks base method.
func (m *MockStorage) GetJobResult(jobID string) (*domain.JobResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJobResult", jobID)
	ret0, _ := ret[0].(*domain.JobResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJobResult indicates an expected call of GetJobResult.
func (mr *MockStorageMockRecorder) GetJobResult(jobID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJobResult", reflect.TypeOf((*MockStorage)(nil).GetJobResult), jobID)
}

// GetJobs mocks base method.
func (m *MockStorage) GetJobs(status domain.JobStatus) ([]*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJobs", status)
	ret0, _ := ret[0].([]*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJobs indicates an expected call of GetJobs.
func (mr *MockStorageMockRecorder) GetJobs(status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJobs", reflect.TypeOf((*MockStorage)(nil).GetJobs), status)
}

// UpdateJob mocks base method.
func (m *MockStorage) UpdateJob(id string, j *domain.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJob", id, j)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJob indicates an expected call of UpdateJob.
func (mr *MockStorageMockRecorder) UpdateJob(id, j interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJob", reflect.TypeOf((*MockStorage)(nil).UpdateJob), id, j)
}

// UpdateJobResult mocks base method.
func (m *MockStorage) UpdateJobResult(jobID string, result *domain.JobResult) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobResult", jobID, result)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobResult indicates an expected call of UpdateJobResult.
func (mr *MockStorageMockRecorder) UpdateJobResult(jobID, result interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobResult", reflect.TypeOf((*MockStorage)(nil).UpdateJobResult), jobID, result)
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
func (m *MockJobQueue) Pop() *domain.Job {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pop")
	ret0, _ := ret[0].(*domain.Job)
	return ret0
}

// Pop indicates an expected call of Pop.
func (mr *MockJobQueueMockRecorder) Pop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pop", reflect.TypeOf((*MockJobQueue)(nil).Pop))
}

// Push mocks base method.
func (m *MockJobQueue) Push(j *domain.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", j)
	ret0, _ := ret[0].(error)
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
func (m *MockJobService) Create(name, taskName, description, runAt string, timeout int, taskParams map[string]interface{}) (*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", name, taskName, description, runAt, timeout, taskParams)
	ret0, _ := ret[0].(*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockJobServiceMockRecorder) Create(name, taskName, description, runAt, timeout, taskParams interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockJobService)(nil).Create), name, taskName, description, runAt, timeout, taskParams)
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

// GetJobs mocks base method.
func (m *MockJobService) GetJobs(status string) ([]*domain.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJobs", status)
	ret0, _ := ret[0].([]*domain.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJobs indicates an expected call of GetJobs.
func (mr *MockJobServiceMockRecorder) GetJobs(status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJobs", reflect.TypeOf((*MockJobService)(nil).GetJobs), status)
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
func (m *MockWorkService) CreateWork(j *domain.Job) work.Work {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWork", j)
	ret0, _ := ret[0].(work.Work)
	return ret0
}

// CreateWork indicates an expected call of CreateWork.
func (mr *MockWorkServiceMockRecorder) CreateWork(j interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWork", reflect.TypeOf((*MockWorkService)(nil).CreateWork), j)
}

// Exec mocks base method.
func (m *MockWorkService) Exec(ctx context.Context, w work.Work) error {
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
func (m *MockWorkService) Send(w work.Work) {
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

// MockConsumer is a mock of Consumer interface.
type MockConsumer struct {
	ctrl     *gomock.Controller
	recorder *MockConsumerMockRecorder
}

// MockConsumerMockRecorder is the mock recorder for MockConsumer.
type MockConsumerMockRecorder struct {
	mock *MockConsumer
}

// NewMockConsumer creates a new mock instance.
func NewMockConsumer(ctrl *gomock.Controller) *MockConsumer {
	mock := &MockConsumer{ctrl: ctrl}
	mock.recorder = &MockConsumerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConsumer) EXPECT() *MockConsumerMockRecorder {
	return m.recorder
}

// Consume mocks base method.
func (m *MockConsumer) Consume(ctx context.Context, duration time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Consume", ctx, duration)
}

// Consume indicates an expected call of Consume.
func (mr *MockConsumerMockRecorder) Consume(ctx, duration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Consume", reflect.TypeOf((*MockConsumer)(nil).Consume), ctx, duration)
}

// MockScheduler is a mock of Scheduler interface.
type MockScheduler struct {
	ctrl     *gomock.Controller
	recorder *MockSchedulerMockRecorder
}

// MockSchedulerMockRecorder is the mock recorder for MockScheduler.
type MockSchedulerMockRecorder struct {
	mock *MockScheduler
}

// NewMockScheduler creates a new mock instance.
func NewMockScheduler(ctrl *gomock.Controller) *MockScheduler {
	mock := &MockScheduler{ctrl: ctrl}
	mock.recorder = &MockSchedulerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScheduler) EXPECT() *MockSchedulerMockRecorder {
	return m.recorder
}

// Schedule mocks base method.
func (m *MockScheduler) Schedule(ctx context.Context, duration time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Schedule", ctx, duration)
}

// Schedule indicates an expected call of Schedule.
func (mr *MockSchedulerMockRecorder) Schedule(ctx, duration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Schedule", reflect.TypeOf((*MockScheduler)(nil).Schedule), ctx, duration)
}

// MockServer is a mock of Server interface.
type MockServer struct {
	ctrl     *gomock.Controller
	recorder *MockServerMockRecorder
}

// MockServerMockRecorder is the mock recorder for MockServer.
type MockServerMockRecorder struct {
	mock *MockServer
}

// NewMockServer creates a new mock instance.
func NewMockServer(ctrl *gomock.Controller) *MockServer {
	mock := &MockServer{ctrl: ctrl}
	mock.recorder = &MockServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServer) EXPECT() *MockServerMockRecorder {
	return m.recorder
}

// GracefullyStop mocks base method.
func (m *MockServer) GracefullyStop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "GracefullyStop")
}

// GracefullyStop indicates an expected call of GracefullyStop.
func (mr *MockServerMockRecorder) GracefullyStop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GracefullyStop", reflect.TypeOf((*MockServer)(nil).GracefullyStop))
}

// Serve mocks base method.
func (m *MockServer) Serve() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Serve")
}

// Serve indicates an expected call of Serve.
func (mr *MockServerMockRecorder) Serve() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Serve", reflect.TypeOf((*MockServer)(nil).Serve))
}
