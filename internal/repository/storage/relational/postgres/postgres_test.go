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

	"github.com/lib/pq"

	"valet/internal/core/domain"
	"valet/internal/repository/storage/relational"
	"valet/pkg/apperrors"
	"valet/pkg/uuidgen"
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
		psqlTest.DB.Close()
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
		t.Fatalf("expected %#v accounts got %#v instead", len(dueJobs), len(dbDueJobs))
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
