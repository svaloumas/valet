package redis

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/mock"
	"github.com/svaloumas/valet/pkg/apperrors"
	"github.com/svaloumas/valet/pkg/uuidgen"

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
				t.Fatalf("expected %#v jobs got %#v instead", len(tt.expected), len(jobs))
			}

			for i := range tt.expected {
				if !reflect.DeepEqual(tt.expected[i], jobs[i]) {
					t.Fatalf("expected %#v got %#v instead", tt.expected[i], jobs[i])
				}
			}
		})
	}
}
func TestRedisGetJobsByPipelineID(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	createdAt := testTime
	runAt := testTime
	job1 := &domain.Job{
		TaskName:  "some task",
		Status:    domain.Pending,
		CreatedAt: &createdAt,
		RunAt:     &runAt,
	}
	createdAt2 := createdAt.Add(1 * time.Minute)
	scheduledAt := testTime.Add(2 * time.Minute)
	job2 := &domain.Job{
		TaskName:    "some other task",
		Status:      domain.Scheduled,
		CreatedAt:   &createdAt2,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}
	createdAt3 := createdAt2.Add(1 * time.Minute)
	startedAt := createdAt.Add(1 * time.Minute)
	job3 := &domain.Job{
		TaskName:    "some task",
		Status:      domain.InProgress,
		CreatedAt:   &createdAt3,
		StartedAt:   &startedAt,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}

	uuid1, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid2, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid3, _ := uuidGenerator.GenerateRandomUUIDString()
	pipelineID, _ := uuidGenerator.GenerateRandomUUIDString()
	notExistingPipelineID, _ := uuidGenerator.GenerateRandomUUIDString()
	job1.ID = uuid1
	job2.ID = uuid2
	job3.ID = uuid3

	job1.NextJobID = job2.ID
	job2.NextJobID = job3.ID

	job1.PipelineID = pipelineID
	job2.PipelineID = pipelineID
	job3.PipelineID = pipelineID

	jobs := []*domain.Job{job1, job2, job3}

	p := &domain.Pipeline{
		ID:          pipelineID,
		Name:        "pipeline_name",
		Description: "some description",
		Jobs:        jobs,
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}
	err := redisTest.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	tests := []struct {
		name       string
		pipelineID string
		expected   []*domain.Job
	}{
		{
			"ok",
			p.ID,
			[]*domain.Job{
				job1,
				job2,
				job3,
			},
		},
		{
			"empty jobs list",
			notExistingPipelineID,
			[]*domain.Job{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			jobs, err := redisTest.GetJobsByPipelineID(tt.pipelineID)
			if err != nil {
				t.Errorf("GetJobsByPipelineID returned unexpected error: got %#v want nil", err)
			}

			if len(tt.expected) != len(jobs) {
				t.Fatalf("expected %#v jobs got %#v instead", len(tt.expected), len(jobs))
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
		t.Fatalf("expected %#v due jobs got %#v instead", len(dueJobs), len(dbDueJobs))
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

func TestRedisCreatePipeline(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	pipelineUUID, _ := uuidGenerator.GenerateRandomUUIDString()
	jobUUID, _ := uuidGenerator.GenerateRandomUUIDString()
	secondJobUUID, _ := uuidGenerator.GenerateRandomUUIDString()

	completedAt := testTime.Add(1 * time.Minute)
	job := &domain.Job{
		ID:          jobUUID,
		Name:        "job_name",
		TaskName:    "test_task",
		PipelineID:  pipelineUUID,
		NextJobID:   secondJobUUID,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: false,
		Timeout:            3,
		Status:             domain.Pending,
		FailureReason:      "",
		RunAt:              &testTime,
		ScheduledAt:        &testTime,
		CreatedAt:          &testTime,
		StartedAt:          &testTime,
		CompletedAt:        &completedAt,
	}

	later := testTime.Add(1 * time.Minute)
	secondJob := &domain.Job{}
	*secondJob = *job
	secondJob.ID = secondJobUUID
	secondJob.Name = "second_job_name"
	secondJob.CreatedAt = &later

	jobs := []*domain.Job{job, secondJob}

	p := &domain.Pipeline{
		ID:          pipelineUUID,
		Name:        "pipeline_name",
		Description: "some description",
		Jobs:        jobs,
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	err := redisTest.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("unexpected error when marshalling pipeline: %#v", err)
	}
	serializedJob, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("unexpected error when marshalling job: %#v", err)
	}
	serializedSecondJob, err := json.Marshal(secondJob)
	if err != nil {
		t.Fatalf("unexpected error when marshalling job: %#v", err)
	}

	pipelineKey := fmt.Sprintf("pipeline:%s", p.ID)
	jobKey := fmt.Sprintf("job:%s", job.ID)
	secondJobKey := fmt.Sprintf("job:%s", secondJob.ID)

	dbJob, err := redisTest.client.Get(ctx, jobKey).Result()
	if err != nil {
		t.Errorf("get job returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedJob), dbJob) {
			t.Fatalf("expected %#v got %#v instead", string(serializedJob), dbJob)
		}
	}

	dbSecondJob, err := redisTest.client.Get(ctx, secondJobKey).Result()
	if err != nil {
		t.Errorf("get job returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedSecondJob), dbSecondJob) {
			t.Fatalf("expected %#v got %#v instead", string(serializedSecondJob), dbSecondJob)
		}
	}

	dbPipeline, err := redisTest.client.Get(ctx, pipelineKey).Result()
	if err != nil {
		t.Errorf("get pipeline returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedPipeline), dbPipeline) {
			t.Fatalf("expected %#v got %#v instead", string(serializedPipeline), dbPipeline)
		}
	}
}

func TestRedisGetPipeline(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	p := &domain.Pipeline{
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	p.ID = uuid

	notExistingPipelineID, _ := uuidGenerator.GenerateRandomUUIDString()

	err := redisTest.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	tests := []struct {
		name     string
		id       string
		pipeline *domain.Pipeline
		err      error
	}{
		{
			"ok",
			p.ID,
			p,
			nil,
		},
		{
			"not found",
			notExistingPipelineID,
			nil,
			&apperrors.NotFoundErr{ID: notExistingPipelineID, ResourceName: "pipeline"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dbPipeline, err := redisTest.GetPipeline(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("GetPipeline returned wrong error: got %#v want %#v", err, tt.err)
				}
			} else {
				if !reflect.DeepEqual(p, dbPipeline) {
					t.Fatalf("expected %#v got %#v instead", p, dbPipeline)
				}
			}
		})
	}
}

func TestRedisGetPipelines(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	createdAt := testTime
	runAt := testTime
	pendingPipeline := &domain.Pipeline{
		Name:      "pending_pipeline",
		Status:    domain.Pending,
		CreatedAt: &createdAt,
		RunAt:     &runAt,
	}
	later := createdAt.Add(1 * time.Minute)
	inprogressPipeline := &domain.Pipeline{
		Name:      "inprogress_pipeline",
		Status:    domain.InProgress,
		CreatedAt: &later,
		RunAt:     &runAt,
		StartedAt: &later,
	}

	uuid1, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid2, _ := uuidGenerator.GenerateRandomUUIDString()
	pendingPipeline.ID = uuid1
	inprogressPipeline.ID = uuid2

	err := redisTest.CreatePipeline(pendingPipeline)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}
	err = redisTest.CreatePipeline(inprogressPipeline)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	tests := []struct {
		name     string
		status   domain.JobStatus
		expected []*domain.Pipeline
	}{
		{
			"all",
			domain.Undefined,
			[]*domain.Pipeline{
				pendingPipeline,
				inprogressPipeline,
			},
		},
		{
			"pending",
			domain.Pending,
			[]*domain.Pipeline{
				pendingPipeline,
			},
		},
		{
			"inprogress",
			domain.InProgress,
			[]*domain.Pipeline{
				inprogressPipeline,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			pipelines, err := redisTest.GetPipelines(tt.status)
			if err != nil {
				t.Errorf("GetPipelines returned unexpected error: got %#v want nil", err)
			}

			if len(tt.expected) != len(pipelines) {
				t.Fatalf("expected %#v pipelines got %#v instead", len(tt.expected), len(pipelines))
			}

			for i := range tt.expected {
				if !reflect.DeepEqual(tt.expected[i], pipelines[i]) {
					t.Fatalf("expected %#v got %#v instead", tt.expected[i], pipelines[i])
				}
			}
		})
	}
}

func TestRedisUpdatePipeline(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	p := &domain.Pipeline{
		Name:        "pipeline_name",
		Description: "some description",
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	p.ID = uuid

	err := redisTest.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	p.Name = "updated_pipeline"
	completedAt := testTime.Add(1 * time.Minute)
	p.CompletedAt = &completedAt

	pipelineKey := fmt.Sprintf("pipeline:%s", p.ID)

	err = redisTest.UpdatePipeline(p.ID, p)
	if err != nil {
		t.Fatalf("unexpected error when updating test pipeline: %#v", err)
	}

	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("unexpected error when marshalling job: %#v", err)
	}

	dbPipeline, err := redisTest.client.Get(ctx, pipelineKey).Result()
	if err != nil {
		t.Errorf("get pipeline returned unexpected error: got %#v want nil", err)
	} else {
		if !reflect.DeepEqual(string(serializedPipeline), dbPipeline) {
			t.Fatalf("expected %#v got %#v instead", string(serializedPipeline), dbPipeline)
		}
	}
}

func TestRedisDeletePipeline(t *testing.T) {
	defer redisTest.client.FlushDB(ctx)

	pipelineUUID, _ := uuidGenerator.GenerateRandomUUIDString()
	jobUUID, _ := uuidGenerator.GenerateRandomUUIDString()
	secondJobUUID, _ := uuidGenerator.GenerateRandomUUIDString()

	completedAt := testTime.Add(1 * time.Minute)
	job := &domain.Job{
		ID:          jobUUID,
		Name:        "job_name",
		TaskName:    "test_task",
		PipelineID:  pipelineUUID,
		NextJobID:   secondJobUUID,
		Description: "some description",
		TaskParams: map[string]interface{}{
			"url": "some-url.com",
		},
		UsePreviousResults: false,
		Timeout:            3,
		Status:             domain.Pending,
		FailureReason:      "",
		RunAt:              &testTime,
		ScheduledAt:        &testTime,
		CreatedAt:          &testTime,
		StartedAt:          &testTime,
		CompletedAt:        &completedAt,
	}

	later := testTime.Add(1 * time.Minute)
	secondJob := &domain.Job{}
	*secondJob = *job
	secondJob.ID = secondJobUUID
	secondJob.Name = "second_job_name"
	secondJob.CreatedAt = &later

	jobs := []*domain.Job{job, secondJob}

	p := &domain.Pipeline{
		ID:          pipelineUUID,
		Name:        "pipeline_name",
		Description: "some description",
		Jobs:        jobs,
		Status:      domain.Pending,
		RunAt:       &testTime,
		CreatedAt:   &testTime,
		StartedAt:   &testTime,
		CompletedAt: &testTime,
	}

	err := redisTest.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	result1 := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}
	result2 := &domain.JobResult{
		JobID:    secondJob.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	pipelineKey := fmt.Sprintf("pipeline:%s", p.ID)
	jobKey := fmt.Sprintf("job:%s", job.ID)
	secondJobKey := fmt.Sprintf("job:%s", secondJob.ID)
	result1Key := fmt.Sprintf("jobresult:%s", result1.JobID)
	result2Key := fmt.Sprintf("jobresult:%s", result2.JobID)

	err = redisTest.CreateJobResult(result1)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	err = redisTest.CreateJobResult(result2)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	err = redisTest.DeletePipeline(p.ID)
	if err != nil {
		t.Fatalf("unexpected error when deleting test pipeline: %#v", err)
	}

	// Should CASCADE
	_, err = redisTest.client.Get(ctx, pipelineKey).Result()
	if err == nil {
		t.Errorf("get pipeline did not return expected error: got %#v want nil", err)
	} else {
		if err != redis.Nil {
			t.Errorf("get pipeline did not return no found error: got %#v want %#v", err, redis.Nil)
		}
	}

	_, err = redisTest.client.Get(ctx, jobKey).Result()
	if err == nil {
		t.Errorf("get job did not return expected error: got %#v want nil", err)
	} else {
		if err != redis.Nil {
			t.Errorf("get job did not return no found error: got %#v want %#v", err, redis.Nil)
		}
	}

	_, err = redisTest.client.Get(ctx, secondJobKey).Result()
	if err == nil {
		t.Errorf("get job did not return expected error: got %#v want nil", err)
	} else {
		if err != redis.Nil {
			t.Errorf("get job did not return no found error: got %#v want %#v", err, redis.Nil)
		}
	}

	_, err = redisTest.client.Get(ctx, result1Key).Result()
	if err == nil {
		t.Errorf("get job result did not return expected error: got %#v want nil", err)
	} else {
		if err != redis.Nil {
			t.Errorf("get job result did not return no found error: got %#v want %#v", err, redis.Nil)
		}
	}

	_, err = redisTest.client.Get(ctx, result2Key).Result()
	if err == nil {
		t.Errorf("get job result did not return expected error: got %#v want nil", err)
	} else {
		if err != redis.Nil {
			t.Errorf("get job result did not return no found error: got %#v want %#v", err, redis.Nil)
		}
	}
}
