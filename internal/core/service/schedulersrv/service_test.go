package schedulersrv

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/service/worksrv/work"
	"github.com/svaloumas/valet/mock"
)

func TestDispatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := new(domain.Job)
	j.ID = "job_1"
	j.TaskName = "test_task"
	j.Status = domain.Pending
	j.RunAt = nil

	w := work.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := &logrus.Logger{Out: ioutil.Discard}

	freezed := mock.NewMockTime(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Pop().
		Return(j).
		Times(1)
	workService := mock.NewMockWorkService(ctrl)
	workService.
		EXPECT().
		Dispatch(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)
	storage := mock.NewMockStorage(ctrl)

	schedulerService := New(jobQueue, storage, workService, freezed, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Dispatch(ctx, 8*time.Millisecond)

	// give some time for the scheduler to consume the job
	time.Sleep(12 * time.Millisecond)
}

func TestDispatchPipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secondJob := &domain.Job{
		ID:         "second_job_id",
		PipelineID: "pipeline_id",
		TaskName:   "some task",
	}

	runAt := time.Now()
	j := &domain.Job{
		ID:         "due_job_id",
		PipelineID: "pipeline_id",
		NextJobID:  secondJob.ID,
		Next:       secondJob,
		TaskName:   "some task",
		RunAt:      &runAt,
	}

	w := work.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := &logrus.Logger{Out: ioutil.Discard}

	freezed := mock.NewMockTime(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Pop().
		Return(j).
		Times(1)
	workService := mock.NewMockWorkService(ctrl)
	workService.
		EXPECT().
		Dispatch(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)
	storage := mock.NewMockStorage(ctrl)

	schedulerService := New(jobQueue, storage, workService, freezed, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Dispatch(ctx, 8*time.Millisecond)

	// give some time for the scheduler to consume the job
	time.Sleep(12 * time.Millisecond)
}

func TestDispatchJobJobInQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := &logrus.Logger{Out: ioutil.Discard}

	freezed := mock.NewMockTime(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)
	jobQueue.
		EXPECT().
		Pop().
		Return(nil).
		Times(2)
	storage := mock.NewMockStorage(ctrl)
	workService := mock.NewMockWorkService(ctrl)

	schedulerService := New(jobQueue, storage, workService, freezed, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Dispatch(ctx, 8*time.Millisecond)

	// give some time for the scheduler to try to consume two jobs
	time.Sleep(20 * time.Millisecond)
}

func TestSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runAt := time.Now()
	j := &domain.Job{
		ID:       "due_job_id",
		RunAt:    &runAt,
		TaskName: "some task",
	}

	w := work.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := &logrus.Logger{Out: ioutil.Discard}

	dueJobs := []*domain.Job{j}

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)
	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetDueJobs().
		Return(dueJobs, nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(j.ID, j).
		Return(nil).
		Times(1)
	workService := mock.NewMockWorkService(ctrl)
	workService.
		EXPECT().
		Dispatch(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)
	jobQueue := mock.NewMockJobQueue(ctrl)

	schedulerService := New(jobQueue, storage, workService, freezed, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Schedule(ctx, 10*time.Millisecond)

	// give some time for the scheduler to schedule the job
	time.Sleep(15 * time.Millisecond)
}

func TestScheduleErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	j := &domain.Job{
		ID:       "due_job_id",
		TaskName: "some task",
	}

	w := work.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := &logrus.Logger{Out: ioutil.Discard}

	dueJobs := []*domain.Job{j}

	storageErr := errors.New("some storage error")

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)

	scheduledAt := freezed.Now()
	j.ScheduledAt = &scheduledAt

	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetDueJobs().
		Return(nil, storageErr).
		Times(1)
	storage.
		EXPECT().
		GetDueJobs().
		Return(dueJobs, nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(j.ID, j).
		Return(storageErr).
		Times(1)
	workService := mock.NewMockWorkService(ctrl)
	workService.
		EXPECT().
		Dispatch(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)
	jobQueue := mock.NewMockJobQueue(ctrl)

	schedulerService := New(jobQueue, storage, workService, freezed, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Schedule(ctx, 10*time.Millisecond)
	// give some time for the scheduler to schedule two jobs
	time.Sleep(25 * time.Millisecond)
}

func TestSchedulePipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secondJob := &domain.Job{
		ID:         "second_job_id",
		PipelineID: "pipeline_id",
		TaskName:   "some task",
	}

	runAt := time.Now()
	j := &domain.Job{
		ID:         "due_job_id",
		PipelineID: "pipeline_id",
		NextJobID:  secondJob.ID,
		Next:       secondJob,
		TaskName:   "some task",
		RunAt:      &runAt,
	}

	w := work.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := &logrus.Logger{Out: ioutil.Discard}

	dueJobs := []*domain.Job{j}

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)
	storage := mock.NewMockStorage(ctrl)
	storage.
		EXPECT().
		GetDueJobs().
		Return(dueJobs, nil).
		Times(1)
	storage.
		EXPECT().
		GetJob(j.NextJobID).
		Return(secondJob, nil).
		Times(1)
	storage.
		EXPECT().
		UpdateJob(j.ID, j).
		Return(nil).
		Times(1)
	workService := mock.NewMockWorkService(ctrl)
	workService.
		EXPECT().
		Dispatch(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)
	jobQueue := mock.NewMockJobQueue(ctrl)

	schedulerService := New(jobQueue, storage, workService, freezed, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Schedule(ctx, 10*time.Millisecond)

	// give some time for the scheduler to schedule the job
	time.Sleep(15 * time.Millisecond)
}
