package schedulersrv

import (
	"context"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"valet/internal/core/domain"
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

	w := domain.Work{
		Job:         j,
		TimeoutUnit: time.Millisecond,
	}
	logger := log.New(ioutil.Discard, "", 0)

	dueJobs := []*domain.Job{j}

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)
	jobRepository := mock.NewMockJobRepository(ctrl)
	jobRepository.
		EXPECT().
		GetDueJobs().
		Return(dueJobs, nil).
		Times(1)
	jobRepository.
		EXPECT().
		Update(j.ID, j).
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

	schedulerService := New(jobRepository, workService, freezed, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerService.Schedule(ctx, 10*time.Millisecond)

	// give some time for the scheduler to schedule the job
	time.Sleep(15 * time.Millisecond)
}

func TestScheduleErrorCases(t *testing.T) {
	// Implement this.
}
