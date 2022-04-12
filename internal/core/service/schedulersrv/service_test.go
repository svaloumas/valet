package schedulersrv

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"valet/internal/core/domain"
	"valet/internal/core/service/worksrv/work"
	"valet/mock"
)

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
		Send(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)

	schedulerService := New(storage, workService, freezed, logger)
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
		Send(w).
		Return().
		Times(1)
	workService.
		EXPECT().
		CreateWork(j).
		Return(w).
		Times(1)

	schedulerService := New(storage, workService, freezed, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Schedule(ctx, 10*time.Millisecond)
	// give some time for the scheduler to schedule two jobs
	time.Sleep(25 * time.Millisecond)
}
