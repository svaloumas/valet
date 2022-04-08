package mysql

import (
	"bytes"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	"valet/internal/core/domain"
	"valet/pkg/apperrors"
	"valet/pkg/uuidgen"
)

const (
	migrationsPath = "./migrations/*.sql"
)

var (
	mysqlTest     *MySQL
	db            *sql.DB
	uuidGenerator uuidgen.UUIDGenerator

	testTime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
)

func TestMain(m *testing.M) {
	mysqlDSN := os.Getenv("MYSQL_DSN")
	mysqlTest = New(mysqlDSN, "", migrationsPath, &MySQLOptions{}, nil)
	uuidGenerator = new(uuidgen.UUIDGen)

	var err error
	db, err = sql.Open("mysql", mysqlTest.Config.FormatDSN())
	if err != nil {
		panic(err)
	}

	defer func() {
		mysqlTest.DB.Close()
		db.Close()
		err := mysqlTest.dropDB()
		if err != nil {
			panic(err)
		}
	}()
	m.Run()
}

func resetDB() {
	var sql bytes.Buffer

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	sql.WriteString("DELETE FROM job;")
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	sql.Reset()
	sql.WriteString("DELETE FROM jobresult;")
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func TestCheckHealth(t *testing.T) {
	result := mysqlTest.CheckHealth()
	if result != true {
		t.Fatalf("expected true got %#v instead", result)
	}
}

func TestMySQLCreateJob(t *testing.T) {
	defer resetDB()

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
		CompletedAt:   &testTime,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	err := mysqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating job: %#v", err)
	}

	dbJob := new(domain.Job)
	var taskParams MapInterface

	var sql bytes.Buffer
	sql.WriteString("SELECT UuidFromBin(id), name, task_name, task_params, ")
	sql.WriteString("timeout, description, status, failure_reason, run_at, ")
	sql.WriteString("scheduled_at, created_at, started_at, completed_at ")
	sql.WriteString("FROM job WHERE id=UuidToBin(?)")

	err = db.QueryRow(sql.String(), job.ID).Scan(
		&dbJob.ID, &dbJob.Name, &dbJob.TaskName, &taskParams, &dbJob.Timeout,
		&dbJob.Description, &dbJob.Status, &dbJob.FailureReason, &dbJob.RunAt,
		&dbJob.ScheduledAt, &dbJob.CreatedAt, &dbJob.StartedAt, &dbJob.CompletedAt)
	if err != nil {
		t.Fatalf("unexpected error when selecting job: %#v", err)
	}

	dbJob.TaskParams = taskParams

	if !reflect.DeepEqual(job, dbJob) {
		t.Fatalf("expected %#v got %#v instead", job, dbJob)
	}
}

func TestMySQLGetJob(t *testing.T) {
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
		CompletedAt:   &testTime,
	}
	uuid, _ := uuidGenerator.GenerateRandomUUIDString()
	job.ID = uuid

	invalidID := "invalid_id"

	err := mysqlTest.CreateJob(job)
	if err != nil {
		t.Fatalf("unexpected error when creating job: %#v", err)
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
			invalidID,
			nil,
			&apperrors.NotFoundErr{ID: invalidID, ResourceName: "job"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dbJob, err := mysqlTest.GetJob(job.ID)
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
