package redis

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
	"valet/internal/core/domain"
	"valet/mock"
	"valet/pkg/apperrors"
	"valet/pkg/uuidgen"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
)

var (
	redisTest     *Redis
	uuidGenerator uuidgen.UUIDGenerator

	testTime   = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	futureTime = time.Date(2029, time.November, 10, 23, 0, 0, 0, time.UTC)
)

func TestMain(m *testing.M) {
	redisURL := os.Getenv("REDIS_URL")
	redisTest = New(redisURL, 1, 5, "")
	uuidGenerator = new(uuidgen.UUIDGen)

	m.Run()
}

func TestCheckHealthRedis(t *testing.T) {
	result := redisTest.CheckHealth()
	if result != true {
		t.Fatalf("expected true got %#v instead", result)
	}
}

func TestRedisCreateJob(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	completedAt := testTime.Add(1 * time.Minute)
	job := &domain.Job{
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Failed,
		FailureReason: "some failure reason",
		RunAt:         &testTime,
		ScheduledAt:   &testTime,
		CreatedAt:     &testTime,
		StartedAt:     &testTime,
		CompletedAt:   &completedAt,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("unexpected error when marshalling job: %#v", err)
	}

	jobKey := fmt.Sprintf("job:%s", job.ID)

	err = redisTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	dbJob, err := redisTest.client.Get(ctx, jobKey).Result()
	if err != nil {
		t.Errorf("get job returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedJob), dbJob) {
			t.Fatalf("expected %#v got %#v instead", string(serializedJob), dbJob)
		}
	}
}

func TestRedisGetJob(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	completedAt := testTime.Add(1 * time.Minute)
	job := &domain.Job{
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Failed,
		FailureReason: "some failure reason",
		RunAt:         &testTime,
		ScheduledAt:   &testTime,
		CreatedAt:     &testTime,
		StartedAt:     &testTime,
		CompletedAt:   &completedAt,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	notExistingJobID, _ := uuidGenerator.GenerateRandomUUIDString()

	err := redisTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	tests := []struct {
		name string
		id   string
		job  *domain.Job
		err  error
	}{
		{
			"ok",
			job.ID,
			job,
			nil,
		},
		{
			"not found",
			notExistingJobID,
			nil,
			&apperrors.NotFoundErr{ID: notExistingJobID, ResourceName: "job"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dbJob, err := redisTest.GetJob(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("GetJob returned wrong error: got %#v want %#v", err, tt.err)
				}
			} else {
				if !reflect.DeepEqual(job, dbJob) {
					t.Fatalf("expected %#v got %#v instead", job, dbJob)
				}
			}
		})
	}
}

func TestRedisGetJobs(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(5)

	createdAt := freezed.Now()
	runAt := freezed.Now()
	pendingJob := &domain.Job{
		ID:        "pending_job_id",
		TaskName:  "some task",
		Status:    domain.Pending,
		CreatedAt: &createdAt,
		RunAt:     &runAt,
	}
	createdAt2 := createdAt.Add(1 * time.Minute)
	scheduledAt := freezed.Now()
	scheduledJob := &domain.Job{
		ID:          "scheduled_job_id",
		TaskName:    "some other task",
		Status:      domain.Scheduled,
		CreatedAt:   &createdAt2,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}
	createdAt3 := createdAt2.Add(1 * time.Minute)
	completedAt := freezed.Now()
	startedAt := freezed.Now()
	inprogressJob := &domain.Job{
		ID:          "inprogress_job_id",
		TaskName:    "some task",
		Status:      domain.InProgress,
		CreatedAt:   &createdAt3,
		StartedAt:   &startedAt,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}
	createdAt4 := createdAt3.Add(1 * time.Minute)
	completedJob := &domain.Job{
		ID:          "completed_job_id",
		TaskName:    "some task",
		Status:      domain.Completed,
		CreatedAt:   &createdAt4,
		StartedAt:   &startedAt,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
		CompletedAt: &completedAt,
	}
	createdAt5 := createdAt4.Add(1 * time.Minute)
	failedJob := &domain.Job{
		ID:            "failed_job_id",
		TaskName:      "some task",
		Status:        domain.Failed,
		FailureReason: "some failure reason",
		CreatedAt:     &createdAt5,
		StartedAt:     &startedAt,
		RunAt:         &runAt,
		ScheduledAt:   &scheduledAt,
		CompletedAt:   &completedAt,
	}

	err := redisTest.CreateJob(pendingJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(scheduledJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(inprogressJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(completedJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(failedJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	tests := []struct {
		name     string
		status   domain.JobStatus
		expected []*domain.Job
	}{
		{
			"all",
			domain.Undefined,
			[]*domain.Job{
				pendingJob,
				scheduledJob,
				inprogressJob,
				completedJob,
				failedJob,
			},
		},
		{
			"pending",
			domain.Pending,
			[]*domain.Job{
				pendingJob,
			},
		},
		{
			"scheduled",
			domain.Scheduled,
			[]*domain.Job{
				scheduledJob,
			},
		},
		{
			"in progress",
			domain.InProgress,
			[]*domain.Job{
				inprogressJob,
			},
		},
		{
			"completed",
			domain.Completed,
			[]*domain.Job{
				completedJob,
			},
		},
		{
			"failed",
			domain.Failed,
			[]*domain.Job{
				failedJob,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			jobs, err := redisTest.GetJobs(tt.status)
			if err != nil {
				t.Errorf("GetJobs returned unexpected error: got %#v want nil", err)
			}

			if len(tt.expected) != len(jobs) {
				t.Fatalf("expected %#v accounts got %#v instead", len(tt.expected), len(jobs))
			}

			for i := range tt.expected {
				if !reflect.DeepEqual(tt.expected[i], jobs[i]) {
					t.Fatalf("expected %#v got %#v instead", tt.expected[i], jobs[i])
				}
			}
		})
	}
}

func TestRedisUpdateJob(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	completedAt := testTime.Add(1 * time.Minute)
	job := &domain.Job{
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Failed,
		FailureReason: "some failure reason",
		RunAt:         &testTime,
		ScheduledAt:   &testTime,
		CreatedAt:     &testTime,
		StartedAt:     &testTime,
		CompletedAt:   &completedAt,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	jobKey := fmt.Sprintf("job:%s", job.ID)

	err := redisTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	job.Name = "updated_name"
	job.Description = "updated description"
	job.Status = 5

	err = redisTest.UpdateJob(job.ID, job)
	if err != nil {
		t.Fatalf("unexpected error when updating test job: %#v", err)
	}

	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("unexpected error when marshalling job: %#v", err)
	}

	dbJob, err := redisTest.client.Get(ctx, jobKey).Result()
	if err != nil {
		t.Errorf("get job returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedJob), dbJob) {
			t.Fatalf("expected %#v got %#v instead", string(serializedJob), dbJob)
		}
	}
}

func TestRedisDeleteJob(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	completedAt := testTime.Add(1 * time.Minute)
	job := &domain.Job{
		Name:        "job_name",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Failed,
		FailureReason: "some failure reason",
		RunAt:         &testTime,
		ScheduledAt:   &testTime,
		CreatedAt:     &testTime,
		StartedAt:     &testTime,
		CompletedAt:   &completedAt,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	jobKey := fmt.Sprintf("job:%s", job.ID)

	err := redisTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	err = redisTest.DeleteJob(job.ID)
	if err != nil {
		t.Fatalf("unexpected error when deleting test job: %#v", err)
	}

	_, err = redisTest.client.Get(ctx, jobKey).Result()
	if err == nil {
		t.Errorf("get job did not return expected error: got %#v want nil", err)
	} else {
		if err != redis.Nil {
			t.Errorf("get job did not return no found error: got %#v want %#v", err, redis.Nil)
		}
	}
}

func TestRedisGetDueJobs(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	duejob1 := &domain.Job{
		Name:        "due_job_1",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Pending,
		FailureReason: "some failure reason",
		RunAt:         &testTime,
		CreatedAt:     &testTime,
	}
	duejob2 := &domain.Job{}
	*duejob2 = *duejob1
	duejob2.Name = "due_job_2"
	aBitLater := testTime.Add(1 * time.Minute)
	duejob2.RunAt = &aBitLater

	notDueJob := &domain.Job{
		Name:        "not_due_job",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Pending,
		FailureReason: "some failure reason",
		RunAt:         &futureTime,
		CreatedAt:     &testTime,
	}

	alreadyScheduledJob := &domain.Job{
		Name:        "already_scheduled_job",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Scheduled,
		FailureReason: "some failure reason",
		RunAt:         &testTime,
		ScheduledAt:   &testTime,
		CreatedAt:     &testTime,
	}

	noScheduleJob := &domain.Job{
		Name:        "no_schedule_job",
		TaskName:    "test_task",
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		Timeout:       3,
		Status:        domain.Pending,
		FailureReason: "some failure reason",
		CreatedAt:     &testTime,
	}

	uuid1, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid2, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid3, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid4, _ := uuidGenerator.GenerateRandomUUIDString()
	duejob1.ID = uuid1
	duejob2.ID = uuid2
	alreadyScheduledJob.ID = uuid3
	noScheduleJob.ID = uuid4

	// ORDER BY run_at ASC
	dueJobs := []*domain.Job{duejob1, duejob2}

	err := redisTest.CreateJob(duejob1)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(duejob2)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(alreadyScheduledJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(notDueJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = redisTest.CreateJob(noScheduleJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	dbDueJobs, err := redisTest.GetDueJobs()
	if err != nil {
		t.Fatalf("unexpected error when getting test due jobs: %#v", err)
	}

	if len(dueJobs) != len(dbDueJobs) {
		t.Fatalf("expected %#v accounts got %#v instead", len(dueJobs), len(dbDueJobs))
	}

	for i := range dueJobs {
		if !reflect.DeepEqual(dueJobs[i], dbDueJobs[i]) {
			t.Fatalf("expected %#v got %#v instead", dueJobs[i], dbDueJobs[i])
		}
	}
}

func TestRedisCreateJobResult(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	result := &domain.JobResult{
		Metadata: "some metadata",
		Error:    "some task error",
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	result.JobID = uuid

	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("unexpected error when marshalling job result: %#v", err)
	}

	jobresultKey := fmt.Sprintf("jobresult:%s", result.JobID)

	err = redisTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	dbJobResult, err := redisTest.client.Get(ctx, jobresultKey).Result()
	if err != nil {
		t.Errorf("get job returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedJobResult), dbJobResult) {
			t.Fatalf("expected %#v got %#v instead", string(serializedJobResult), dbJobResult)
		}
	}
}

func TestRedisGetJobResult(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	result := &domain.JobResult{
		Metadata: "some metadata",
		Error:    "some task error",
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	result.JobID = uuid

	notExistingJobResultID, _ := uuidGenerator.GenerateRandomUUIDString()

	err := redisTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	tests := []struct {
		name   string
		id     string
		result *domain.JobResult
		err    error
	}{
		{
			"ok",
			result.JobID,
			result,
			nil,
		},
		{
			"not found",
			notExistingJobResultID,
			nil,
			&apperrors.NotFoundErr{ID: notExistingJobResultID, ResourceName: "job result"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dbJobResult, err := redisTest.GetJobResult(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("GetJob returned wrong error: got %#v want %#v", err, tt.err)
				}
			} else {
				if !reflect.DeepEqual(result, dbJobResult) {
					t.Fatalf("expected %#v got %#v instead", result, dbJobResult)
				}
			}
		})
	}
}

func TestRedisUpdateJobResult(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	result := &domain.JobResult{
		Metadata: "some metadata",
		Error:    "some task error",
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	result.JobID = uuid

	jobResultKey := fmt.Sprintf("jobresult:%s", result.JobID)

	err := redisTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	result.Metadata = "updated metadata"
	result.Error = "updated task error"

	err = redisTest.UpdateJobResult(result.JobID, result)
	if err != nil {
		t.Fatalf("unexpected error when updating test job result: %#v", err)
	}

	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("unexpected error when marshalling job result: %#v", err)
	}

	dbJobResult, err := redisTest.client.Get(ctx, jobResultKey).Result()
	if err != nil {
		t.Errorf("get job returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedJobResult), dbJobResult) {
			t.Fatalf("expected %#v got %#v instead", string(serializedJobResult), dbJobResult)
		}
	}
}

func TestRedisDeleteJobResult(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	result := &domain.JobResult{
		Metadata: "some metadata",
		Error:    "some task error",
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	result.JobID = uuid

	jobResultKey := fmt.Sprintf("jobresult:%s", result.JobID)

	err := redisTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	err = redisTest.DeleteJobResult(result.JobID)
	if err != nil {
		t.Fatalf("unexpected error when deleting test job result: %#v", err)
	}

	_, err = redisTest.client.Get(ctx, jobResultKey).Result()
	if err == nil {
		t.Errorf("get job did not return expected error: got %#v want nil", err)
	} else {
		if err != redis.Nil {
			t.Errorf("get job did not return no found error: got %#v want %#v", err, redis.Nil)
		}
	}
}
