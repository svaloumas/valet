package postgres

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lib/pq"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/repository/storage/relational"
	"github.com/svaloumas/valet/pkg/apperrors"
	"github.com/svaloumas/valet/pkg/uuidgen"
)

var (
	psqlTest      *PostgreSQL
	psqlDSN       string
	uuidGenerator uuidgen.UUIDGenerator
	dbName        string

	testTime   = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	futureTime = time.Date(2029, time.November, 10, 23, 0, 0, 0, time.UTC)
)

func TestMain(m *testing.M) {
	psqlDSN = os.Getenv("POSTGRES_DSN")
	psqlTest = New(psqlDSN, &relational.DBOptions{})
	uuidGenerator = new(uuidgen.UUIDGen)

	pgDSN, err := pq.ParseURL(psqlDSN)
	if err != nil {
		panic(err)
	}

	parts := strings.Split(pgDSN, " ")
	if len(parts) < 5 {
		panic("missing parts in the postgres connection string")
	}

	var dbNamePart string
	found := false
	for _, part := range parts {
		if strings.Contains(part, "dbname=") {
			dbNamePart = part
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("invalid postgres connection string: %s", pgDSN))
	}
	quotedDBName := strings.Split(dbNamePart, "=")[1]
	dbName = strings.Replace(quotedDBName, "'", "", 2)

	createDB(psqlDSN, dbName)
	defer func() {
		dropTestDB()
		psqlTest.Close()
	}()
	m.Run()
}

func resetTestDB() {
	var sql bytes.Buffer

	tx, err := psqlTest.DB.Begin()
	if err != nil {
		panic(err)
	}

	sql.WriteString("DELETE FROM jobresult;")
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	sql.Reset()
	sql.WriteString("DELETE FROM job;")
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	sql.Reset()
	sql.WriteString("DELETE FROM pipeline;")
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func dropTestDB() {
	dsnWithoutDBName := strings.Replace(psqlDSN, dbName, "", 1)

	db, err := sql.Open("postgres", dsnWithoutDBName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var query bytes.Buffer
	query.WriteString(fmt.Sprintf("DROP DATABASE %s", dbName))
	_, err = db.Exec(query.String())
	if err != nil {
		panic(err)
	}
}
func TestPanicWithInvalidDSN(t *testing.T) {
	psqlDSN := "invalid_dsn"
	defer func() {
		if p := recover(); p == nil {
			t.Errorf("New did not panic with invalid URL")
		} else {
			panicMsg := "invalid connection protocol: "
			if err := fmt.Errorf("%s", p); err.Error() != panicMsg {
				t.Errorf("New paniced with unexpected panic message: got %v want %v", err.Error(), panicMsg)
			}
		}
	}()
	New(psqlDSN, &relational.DBOptions{})
}

func TestCheckHealth(t *testing.T) {
	result := psqlTest.CheckHealth()
	if result != true {
		t.Fatalf("expected true got %#v instead", result)
	}
}

func TestPostgreSQLCreateJob(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	dbJob := new(domain.Job)
	var taskParams relational.MapStringInterface

	var sql bytes.Buffer
	sql.WriteString("SELECT id, name, task_name, task_params, ")
	sql.WriteString("timeout, description, status, failure_reason, run_at, ")
	sql.WriteString("scheduled_at, created_at, started_at, completed_at ")
	sql.WriteString("FROM job WHERE id=$1")

	err = psqlTest.DB.QueryRow(sql.String(), job.ID).Scan(
		&dbJob.ID, &dbJob.Name, &dbJob.TaskName, &taskParams, &dbJob.Timeout,
		&dbJob.Description, &dbJob.Status, &dbJob.FailureReason, &dbJob.RunAt,
		&dbJob.ScheduledAt, &dbJob.CreatedAt, &dbJob.StartedAt, &dbJob.CompletedAt)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job: %#v", err)
	}

	dbJob.TaskParams = taskParams

	if !reflect.DeepEqual(job, dbJob) {
		t.Fatalf("expected %#v got %#v instead", job, dbJob)
	}
}

func TestPostgreSQLGetJob(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreateJob(job)
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
			uuid,
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

			dbJob, err := psqlTest.GetJob(tt.id)
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

func TestPostgreSQLGetJobs(t *testing.T) {
	defer resetTestDB()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createdAt := testTime
	runAt := testTime
	pendingJob := &domain.Job{
		TaskName:  "some task",
		Status:    domain.Pending,
		CreatedAt: &createdAt,
		RunAt:     &runAt,
	}
	createdAt2 := createdAt.Add(1 * time.Minute)
	scheduledAt := testTime.Add(2 * time.Minute)
	scheduledJob := &domain.Job{
		TaskName:    "some other task",
		Status:      domain.Scheduled,
		CreatedAt:   &createdAt2,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}
	createdAt3 := createdAt2.Add(1 * time.Minute)
	completedAt := createdAt3.Add(5 * time.Minute)
	startedAt := createdAt.Add(1 * time.Minute)
	inprogressJob := &domain.Job{
		TaskName:    "some task",
		Status:      domain.InProgress,
		CreatedAt:   &createdAt3,
		StartedAt:   &startedAt,
		RunAt:       &runAt,
		ScheduledAt: &scheduledAt,
	}
	createdAt4 := createdAt3.Add(1 * time.Minute)
	completedJob := &domain.Job{
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
		TaskName:      "some task",
		Status:        domain.Failed,
		FailureReason: "some failure reason",
		CreatedAt:     &createdAt5,
		StartedAt:     &startedAt,
		RunAt:         &runAt,
		ScheduledAt:   &scheduledAt,
		CompletedAt:   &completedAt,
	}

	uuid1, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid2, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid3, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid4, _ := uuidGenerator.GenerateRandomUUIDString()
	uuid5, _ := uuidGenerator.GenerateRandomUUIDString()
	pendingJob.ID = uuid1
	scheduledJob.ID = uuid2
	inprogressJob.ID = uuid3
	completedJob.ID = uuid4
	failedJob.ID = uuid5

	err := psqlTest.CreateJob(pendingJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(scheduledJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(inprogressJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(completedJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(failedJob)
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

			jobs, err := psqlTest.GetJobs(tt.status)
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

func TestPostgreSQLGetJobsByPipelineID(t *testing.T) {
	defer resetTestDB()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	err := psqlTest.CreateJob(job1)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(job2)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(job3)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	tests := []struct {
		name       string
		pipelineID string
		expected   []*domain.Job
	}{
		{
			"ok",
			pipelineID,
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

			jobs, err := psqlTest.GetJobsByPipelineID(tt.pipelineID)
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

func TestPostgreSQLUpdateJob(t *testing.T) {
	defer resetTestDB()

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
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	err := psqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	job.Name = "test_job"
	job.TaskName = "another_task"
	job.TaskParams = map[string]interface{}{
		"addr": ":8000",
	}
	completedAt := testTime.Add(1 * time.Minute)
	job.CompletedAt = &completedAt

	err = psqlTest.UpdateJob(job.ID, job)
	if err != nil {
		t.Errorf("unexpected error when updating test job: %#v", err)
	}

	dbJob := new(domain.Job)
	var taskParams relational.MapStringInterface

	var sql bytes.Buffer
	sql.WriteString("SELECT id, name, task_name, task_params, ")
	sql.WriteString("timeout, description, status, failure_reason, run_at, ")
	sql.WriteString("scheduled_at, created_at, started_at, completed_at ")
	sql.WriteString("FROM job WHERE id=$1")

	err = psqlTest.DB.QueryRow(sql.String(), job.ID).Scan(
		&dbJob.ID, &dbJob.Name, &dbJob.TaskName, &taskParams, &dbJob.Timeout,
		&dbJob.Description, &dbJob.Status, &dbJob.FailureReason, &dbJob.RunAt,
		&dbJob.ScheduledAt, &dbJob.CreatedAt, &dbJob.StartedAt, &dbJob.CompletedAt)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job: %#v", err)
	}

	dbJob.TaskParams = taskParams

	if !reflect.DeepEqual(job, dbJob) {
		t.Fatalf("expected %#v got %#v instead", job, dbJob)
	}
}

func TestPostgreSQLDeleteJob(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	result := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	err = psqlTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	err = psqlTest.DeleteJob(job.ID)
	if err != nil {
		t.Fatalf("unexpected error when deleting test job: %#v", err)
	}

	var sql bytes.Buffer
	sql.WriteString("SELECT * FROM job WHERE id=$1")
	rows, err := psqlTest.DB.Query(sql.String(), job.ID)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job: %#v", err)
	}
	if rows.Next() {
		t.Fatal("expected the job row to be deleted, got some back")
	}
	// Should CASCADE
	sql.Reset()
	sql.WriteString("SELECT * FROM jobresult WHERE job_id=$1")
	rows, err = psqlTest.DB.Query(sql.String(), job.ID)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job result: %#v", err)
	}
	if rows.Next() {
		t.Fatal("expected the job row to be deleted, got some back")
	}
}

func TestPostgreSQLGetDueJobs(t *testing.T) {
	defer resetTestDB()

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
	uuid5, _ := uuidGenerator.GenerateRandomUUIDString()
	duejob1.ID = uuid1
	duejob2.ID = uuid2
	alreadyScheduledJob.ID = uuid3
	noScheduleJob.ID = uuid4
	notDueJob.ID = uuid5

	// ORDER BY run_at ASC
	dueJobs := []*domain.Job{duejob1, duejob2}

	err := psqlTest.CreateJob(duejob1)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(duejob2)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(alreadyScheduledJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(notDueJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}
	err = psqlTest.CreateJob(noScheduleJob)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	dbDueJobs, err := psqlTest.GetDueJobs()
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

func TestPostgreSQLCreateJobResult(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	result := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	err = psqlTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	var metadataBytes []byte
	dbResult := new(domain.JobResult)

	var sql bytes.Buffer
	sql.WriteString("SELECT job_id, metadata, error ")
	sql.WriteString("FROM jobresult WHERE job_id=$1")

	err = psqlTest.DB.QueryRow(sql.String(), result.JobID).Scan(
		&dbResult.JobID, &metadataBytes, &dbResult.Error)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job result: %#v", err)
	}
	err = json.Unmarshal(metadataBytes, &dbResult.Metadata)
	if err != nil {
		t.Fatalf("unexpected error when unmarshaling test job result metadata: %#v", err)
	}

	if !reflect.DeepEqual(result, dbResult) {
		t.Fatalf("expected %#v got %#v instead", result, dbResult)
	}
}

func TestPostgreSQLGetJobResult(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	result := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	err = psqlTest.CreateJobResult(result)
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
			uuid,
			result,
			nil,
		},
		{
			"not found",
			notExistingJobID,
			nil,
			&apperrors.NotFoundErr{ID: notExistingJobID, ResourceName: "job result"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dbJob, err := psqlTest.GetJobResult(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("GetJobResult returned wrong error: got %#v want %#v", err, tt.err)
				}
			} else {
				if !reflect.DeepEqual(result, dbJob) {
					t.Fatalf("expected %#v got %#v instead", job, dbJob)
				}
			}
		})
	}
}

func TestPostgreSQLUpdateJobResult(t *testing.T) {
	defer resetTestDB()

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
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	err := psqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	result := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	err = psqlTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	result.Metadata = "updated_metadata"
	result.Error = "updated_error"

	err = psqlTest.UpdateJobResult(result.JobID, result)
	if err != nil {
		t.Errorf("unexpected error when updating test job result: %#v", err)
	}

	var metadataBytes []byte
	dbResult := new(domain.JobResult)

	var sql bytes.Buffer
	sql.WriteString("SELECT job_id, metadata, error ")
	sql.WriteString("FROM jobresult WHERE job_id=$1")

	err = psqlTest.DB.QueryRow(sql.String(), job.ID).Scan(
		&dbResult.JobID, &metadataBytes, &dbResult.Error)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job result: %#v", err)
	}

	err = json.Unmarshal(metadataBytes, &dbResult.Metadata)
	if err != nil {
		t.Fatalf("unexpected error when unmarshaling test job result metadata: %#v", err)
	}

	if !reflect.DeepEqual(result, dbResult) {
		t.Fatalf("expected %#v got %#v instead", result, dbResult)
	}
}

func TestPostgreSQLDeleteJobResult(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating test job: %#v", err)
	}

	result := &domain.JobResult{
		JobID:    job.ID,
		Metadata: "some metadata",
		Error:    "some task error",
	}

	err = psqlTest.CreateJobResult(result)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	err = psqlTest.DeleteJobResult(result.JobID)
	if err != nil {
		t.Fatalf("unexpected error when deleting test job result: %#v", err)
	}

	var sql bytes.Buffer
	sql.WriteString("SELECT * FROM jobresult WHERE job_id=$1")
	rows, err := psqlTest.DB.Query(sql.String(), job.ID)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job result: %#v", err)
	}
	if rows.Next() {
		t.Fatal("expected the job row to be deleted, got some back")
	}
}

func TestPostgreSQLCreatePipeline(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	dbPipeline := new(domain.Pipeline)

	var query bytes.Buffer
	query.WriteString("SELECT id, name, description, status, run_at, ")
	query.WriteString("created_at, started_at, completed_at ")
	query.WriteString("FROM pipeline WHERE id=$1")

	err = psqlTest.DB.QueryRow(query.String(), p.ID).Scan(
		&dbPipeline.ID, &dbPipeline.Name, &dbPipeline.Description, &dbPipeline.Status,
		&dbPipeline.RunAt, &dbPipeline.CreatedAt, &dbPipeline.StartedAt, &dbPipeline.CompletedAt)
	if err != nil {
		t.Fatalf("unexpected error when selecting test pipeline: %#v", err)
	}

	query.Reset()
	query.WriteString("SELECT id, pipeline_id, next_job_id, ")
	query.WriteString("use_previous_results, name, task_name, task_params, timeout, description, status, ")
	query.WriteString("failure_reason, run_at, scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job WHERE pipeline_id=$1 ORDER BY created_at ASC")

	rows, err := psqlTest.DB.Query(query.String(), p.ID)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("unexpected error when selecting test pipeline jobs: %#v", err)
	}

	dbJobs := make([]*domain.Job, 0)
	for rows.Next() {
		var taskParams relational.MapStringInterface
		dbJob := new(domain.Job)
		err = rows.Scan(&dbJob.ID, &dbJob.PipelineID, &dbJob.NextJobID, &dbJob.UsePreviousResults, &dbJob.Name,
			&dbJob.TaskName, &taskParams, &dbJob.Timeout, &dbJob.Description, &dbJob.Status, &dbJob.FailureReason,
			&dbJob.RunAt, &dbJob.ScheduledAt, &dbJob.CreatedAt, &dbJob.StartedAt, &dbJob.CompletedAt)
		if err != nil {
			t.Fatalf("unexpected error when selecting test pipeline job: %#v", err)
		}
		dbJob.TaskParams = taskParams

		dbJobs = append(dbJobs, dbJob)
	}

	for i, j := range p.Jobs {
		if !reflect.DeepEqual(j, dbJobs[i]) {
			t.Fatalf("expected %#v got %#v instead", j, dbJobs[i])
		}
	}

	// Already checked for jobs due to reflect.DeepEqual on slice.
	p.Jobs = nil
	if !reflect.DeepEqual(p, dbPipeline) {
		t.Fatalf("expected %#v got %#v instead", p, dbPipeline)
	}
}

func TestPostgreSQLGetPipeline(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreatePipeline(p)
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
			uuid,
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

			dbPipeline, err := psqlTest.GetPipeline(tt.id)
			if err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("GetJob returned wrong error: got %#v want %#v", err, tt.err)
				}
			} else {
				if !reflect.DeepEqual(p, dbPipeline) {
					t.Fatalf("expected %#v got %#v instead", p, dbPipeline)
				}
			}
		})
	}
}

func TestPostgreSQLGetPipelines(t *testing.T) {
	defer resetTestDB()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	err := psqlTest.CreatePipeline(pendingPipeline)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}
	err = psqlTest.CreatePipeline(inprogressPipeline)
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

			pipelines, err := psqlTest.GetPipelines(tt.status)
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

func TestPostgreSQLUpdatePipeline(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreatePipeline(p)
	if err != nil {
		t.Fatalf("unexpected error when creating test pipeline: %#v", err)
	}

	p.Name = "updated_pipeline"
	completedAt := testTime.Add(1 * time.Minute)
	p.CompletedAt = &completedAt

	err = psqlTest.UpdatePipeline(p.ID, p)
	if err != nil {
		t.Errorf("unexpected error when updating test pipeline: %#v", err)
	}

	dbPipeline := new(domain.Pipeline)

	var query bytes.Buffer
	query.WriteString("SELECT id, name, description, status, run_at, ")
	query.WriteString("created_at, started_at, completed_at ")
	query.WriteString("FROM pipeline WHERE id=$1")

	err = psqlTest.DB.QueryRow(query.String(), p.ID).Scan(
		&dbPipeline.ID, &dbPipeline.Name, &dbPipeline.Description, &dbPipeline.Status,
		&dbPipeline.RunAt, &dbPipeline.CreatedAt, &dbPipeline.StartedAt, &dbPipeline.CompletedAt)
	if err != nil {
		t.Fatalf("unexpected error when selecting test pipeline: %#v", err)
	}

	if !reflect.DeepEqual(p, dbPipeline) {
		t.Fatalf("expected %#v got %#v instead", p, dbPipeline)
	}
}

func TestPostgreSQLDeletePipeline(t *testing.T) {
	defer resetTestDB()

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

	err := psqlTest.CreatePipeline(p)
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

	err = psqlTest.CreateJobResult(result1)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	err = psqlTest.CreateJobResult(result2)
	if err != nil {
		t.Fatalf("unexpected error when creating test job result: %#v", err)
	}

	err = psqlTest.DeletePipeline(p.ID)
	if err != nil {
		t.Fatalf("unexpected error when deleting test pipeline: %#v", err)
	}

	var sql bytes.Buffer
	sql.WriteString("SELECT * FROM job WHERE pipeline_id=$1")
	rows, err := psqlTest.DB.Query(sql.String(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job: %#v", err)
	}
	if rows.Next() {
		t.Fatal("expected the job row to be deleted, got some back")
	}
	// Should CASCADE
	sql.Reset()
	sql.WriteString("SELECT * FROM jobresult WHERE job_id=$1")
	rows, err = psqlTest.DB.Query(sql.String(), job.ID)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job result: %#v", err)
	}
	if rows.Next() {
		t.Fatal("expected the job result row to be deleted, got some back")
	}

	sql.Reset()
	sql.WriteString("SELECT * FROM jobresult WHERE job_id=$1")
	rows, err = psqlTest.DB.Query(sql.String(), secondJob.ID)
	if err != nil {
		t.Fatalf("unexpected error when selecting test job result: %#v", err)
	}
	if rows.Next() {
		t.Fatal("expected the job row to be deleted, got some back")
	}

	sql.Reset()
	sql.WriteString("SELECT * FROM pipeline WHERE id=$1")
	rows, err = psqlTest.DB.Query(sql.String(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error when selecting test pipeline: %#v", err)
	}
	if rows.Next() {
		t.Fatal("expected the pipeline row to be deleted, got some back")
	}
}
