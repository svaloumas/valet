package postgres

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/repository/storage/relational"
	"github.com/svaloumas/valet/pkg/apperrors"
)

const (
	createPipelineTableMigration = `
		CREATE TABLE IF NOT EXISTS pipeline (
		id UUID NOT NULL,
		name varchar(255) NOT NULL,
		description varchar(255) NOT NULL DEFAULT '',
		status INT NOT NULL DEFAULT 0,
		run_at timestamp with time zone NULL,
		created_at timestamp with time zone default (now() at time zone 'utc'),
		started_at timestamp with time zone NULL,
		completed_at timestamp with time zone NULL,
		PRIMARY KEY (id)
		);
	`
	createJobTableMigration = `
		CREATE TABLE IF NOT EXISTS job (
		id UUID NOT NULL,
		pipeline_id varchar(36) NOT NULL default '',
		next_job_id varchar(36) NOT NULL default '',
		use_previous_results BOOLEAN default false,
		name varchar(255) NOT NULL,
		task_name varchar(255) NOT NULL,
		task_params JSON NOT NULL,
		timeout INT NOT NULL,
		description varchar(255) NOT NULL DEFAULT '',
		status INT NOT NULL DEFAULT 0,
		failure_reason TEXT NOT NULL,
		run_at timestamp with time zone NULL,
		scheduled_at timestamp with time zone NULL,
		created_at timestamp with time zone default (now() at time zone 'utc'),
		started_at timestamp with time zone NULL,
		completed_at timestamp with time zone NULL,
		PRIMARY KEY (id)
		);
	`
	createJobResultTableMigration = `
		CREATE TABLE IF NOT EXISTS jobresult (
		job_id UUID NOT NULL,
		metadata JSON NOT NULL,
		error TEXT NOT NULL,
		PRIMARY KEY (job_id),
		CONSTRAINT fk_job_id FOREIGN KEY (job_id) REFERENCES job (id)
		);
	`
	createIndexes = `
	    CREATE INDEX IF NOT EXISTS idx_job_status ON job(status);
	    CREATE INDEX IF NOT EXISTS idx_job_pipeline_id ON job(pipeline_id);
	    CREATE INDEX IF NOT EXISTS idx_pipeline_status ON pipeline(status);
	`
)

var _ port.Storage = &PostgreSQL{}

type PostgreSQL struct {
	DB *sql.DB
}

// New initializes and returns a PostgreSQL client.
func New(dsn string, options *relational.DBOptions) *PostgreSQL {
	psql := new(PostgreSQL)

	pgDSN, err := pq.ParseURL(dsn)
	if err != nil {
		panic(err)
	}

	parts := strings.Split(pgDSN, " ")
	if len(parts) < 5 {
		panic("missing parts in the postgres connection string")
	}

	// TODO: Revisit this.
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
		panic("invalid Postgres connection string")
	}
	quotedDBName := strings.Split(dbNamePart, "=")[1]
	dbName := strings.Replace(quotedDBName, "'", "", 2)

	createDB(dsn, dbName)

	psql.DB, err = sql.Open("postgres", pgDSN)
	if err != nil {
		panic(err)
	}
	psql.DB.SetConnMaxLifetime(time.Duration(options.ConnectionMaxLifetime) * time.Millisecond)
	psql.DB.SetMaxIdleConns(options.MaxIdleConnections)
	psql.DB.SetMaxOpenConns(options.MaxOpenConnections)
	return psql
}

// CheckHealth returns the status of MySQL.
func (psql *PostgreSQL) CheckHealth() bool {
	err := psql.DB.Ping()
	return err == nil
}

// Close terminates any storage connections gracefully.
func (psql *PostgreSQL) Close() error {
	return psql.DB.Close()
}

// CreateJob adds a new job to the storage.
func (psql *PostgreSQL) CreateJob(j *domain.Job) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO job (id, name, pipeline_id, next_job_id, task_name, task_params, ")
	query.WriteString("use_previous_results, timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at) ")
	query.WriteString("VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)")

	var taskParams relational.MapStringInterface = j.TaskParams
	res, err := tx.Exec(query.String(), j.ID, j.Name, j.PipelineID, j.NextJobID, j.TaskName, taskParams,
		j.UsePreviousResults, j.Timeout, j.Description, j.Status, j.FailureReason, j.RunAt,
		j.ScheduledAt, j.CreatedAt, j.StartedAt, j.CompletedAt)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		tx.Rollback()
		return fmt.Errorf("could not insert job, rows affected: %d", rowsAffected)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetJob fetches a job from the storage.
func (psql *PostgreSQL) GetJob(id string) (*domain.Job, error) {
	var query bytes.Buffer
	query.WriteString("SELECT id, name, pipeline_id, next_job_id, task_name, task_params, ")
	query.WriteString("use_previous_results, timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job WHERE id=$1")

	var taskParams relational.MapStringInterface
	job := new(domain.Job)

	err := psql.DB.QueryRow(query.String(), id).Scan(
		&job.ID, &job.Name, &job.PipelineID, &job.NextJobID, &job.TaskName, &taskParams, &job.UsePreviousResults,
		&job.Timeout, &job.Description, &job.Status, &job.FailureReason, &job.RunAt,
		&job.ScheduledAt, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
	}
	job.TaskParams = taskParams
	return job, nil
}

// GetJobs fetches all jobs from the storage, optionally filters the jobs by status.
func (psql *PostgreSQL) GetJobs(status domain.JobStatus) ([]*domain.Job, error) {
	filterByStatus := ""
	if status != domain.Undefined {
		filterByStatus = fmt.Sprintf("WHERE status = %d", status.Index())
	}

	var query bytes.Buffer
	query.WriteString("SELECT id, name, pipeline_id, next_job_id, task_name, task_params, ")
	query.WriteString("use_previous_results, timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job " + filterByStatus + " ORDER BY created_at ASC")

	jobs := make([]*domain.Job, 0)

	rows, err := psql.DB.Query(query.String())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var taskParams relational.MapStringInterface
		job := new(domain.Job)
		err = rows.Scan(&job.ID, &job.Name, &job.PipelineID, &job.NextJobID, &job.TaskName, &taskParams,
			&job.UsePreviousResults, &job.Timeout, &job.Description, &job.Status, &job.FailureReason,
			&job.RunAt, &job.ScheduledAt, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// UpdateJob updates a job to the storage.
func (psql *PostgreSQL) UpdateJob(id string, j *domain.Job) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}
	var query bytes.Buffer
	var taskParams relational.MapStringInterface = j.TaskParams

	query.WriteString("UPDATE job SET name=$1, task_name=$2, task_params=$3, timeout=$4, ")
	query.WriteString("description=$5, status=$6, failure_reason=$7, run_at=$8, ")
	query.WriteString("scheduled_at=$9, created_at=$10, started_at=$11, completed_at=$12 ")
	query.WriteString("WHERE id=$13")

	res, err := tx.Exec(query.String(), j.Name, j.TaskName, taskParams,
		j.Timeout, j.Description, j.Status, j.FailureReason, j.RunAt,
		j.ScheduledAt, j.CreatedAt, j.StartedAt, j.CompletedAt, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		tx.Rollback()
		return fmt.Errorf("could not update job, rows affected: %d", rowsAffected)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeleteJob deletes a job from the storage.
func (psql *PostgreSQL) DeleteJob(id string) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}

	// CASCADE
	var query bytes.Buffer
	query.WriteString("DELETE FROM jobresult WHERE job_id=$1")
	if _, err = tx.Exec(query.String(), id); err != nil {
		tx.Rollback()
		return err
	}

	query.Reset()
	query.WriteString("DELETE FROM job WHERE id=$1")
	if _, err = tx.Exec(query.String(), id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (psql *PostgreSQL) GetDueJobs() ([]*domain.Job, error) {

	var query bytes.Buffer
	query.WriteString("SELECT id, name, pipeline_id, next_job_id, task_name, task_params, ")
	query.WriteString("use_previous_results, timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job WHERE run_at IS NOT NULL AND run_at < $1 ")
	query.WriteString("AND status = 1 ORDER BY run_at ASC")

	dueJobs := make([]*domain.Job, 0)

	rows, err := psql.DB.Query(query.String(), time.Now())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var taskParams relational.MapStringInterface
		job := new(domain.Job)

		err := rows.Scan(
			&job.ID, &job.Name, &job.PipelineID, &job.NextJobID, &job.TaskName, &taskParams,
			&job.UsePreviousResults, &job.Timeout, &job.Description, &job.Status, &job.FailureReason,
			&job.RunAt, &job.ScheduledAt, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
		if err != nil {
			return nil, err
		}
		job.TaskParams = taskParams
		dueJobs = append(dueJobs, job)
	}

	return dueJobs, nil
}

// GetJobsByPipelineID fetches the jobs of the specified pipeline.
func (psql *PostgreSQL) GetJobsByPipelineID(pipelineID string) ([]*domain.Job, error) {
	var query bytes.Buffer
	query.WriteString("SELECT id, pipeline_id, next_job_id, ")
	query.WriteString("use_previous_results, name, task_name, task_params, timeout, description, status, ")
	query.WriteString("failure_reason, run_at, scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job WHERE pipeline_id=$1 ORDER BY created_at ASC")

	jobs := make([]*domain.Job, 0)

	rows, err := psql.DB.Query(query.String(), pipelineID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var taskParams relational.MapStringInterface
		job := new(domain.Job)
		err = rows.Scan(&job.ID, &job.PipelineID, &job.NextJobID, &job.UsePreviousResults, &job.Name,
			&job.TaskName, &taskParams, &job.Timeout, &job.Description, &job.Status, &job.FailureReason,
			&job.RunAt, &job.ScheduledAt, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// CreateJobResult adds new job result to the storage.
func (psql *PostgreSQL) CreateJobResult(result *domain.JobResult) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}

	metadataBytes, err := json.Marshal(result.Metadata)
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO jobresult (job_id, metadata, error) ")
	query.WriteString("VALUES ($1, $2, $3)")

	res, err := tx.Exec(query.String(), result.JobID, metadataBytes, result.Error)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		tx.Rollback()
		return fmt.Errorf("could not insert job result, rows affected: %d", rowsAffected)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetJobResult fetches a job result from the storage.
func (psql *PostgreSQL) GetJobResult(jobID string) (*domain.JobResult, error) {
	var query bytes.Buffer
	query.WriteString("SELECT job_id, metadata, error ")
	query.WriteString("FROM jobresult WHERE job_id=$1")

	var metadataBytes []byte
	result := new(domain.JobResult)

	err := psql.DB.QueryRow(query.String(), jobID).Scan(
		&result.JobID, &metadataBytes, &result.Error)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, &apperrors.NotFoundErr{ID: jobID, ResourceName: "job result"}
	}
	if err := json.Unmarshal(metadataBytes, &result.Metadata); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateJobResult updates a job result to the storage.
func (psql *PostgreSQL) UpdateJobResult(jobID string, result *domain.JobResult) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}
	metadataBytes, err := json.Marshal(result.Metadata)
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("UPDATE jobresult SET metadata=$1, error=$2 ")
	query.WriteString("WHERE job_id=$3")

	res, err := tx.Exec(query.String(), metadataBytes, result.Error, jobID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		tx.Rollback()
		return fmt.Errorf("could not update job result, rows affected: %d", rowsAffected)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeleteJobResult deletes a job result from the storage.
func (psql *PostgreSQL) DeleteJobResult(jobID string) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("DELETE FROM jobresult WHERE job_id=$1")
	if _, err = tx.Exec(query.String(), jobID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// CreatePipeline adds a new pipeline and of its jobs to the storage.
func (psql *PostgreSQL) CreatePipeline(p *domain.Pipeline) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO pipeline (id, name, description, ")
	query.WriteString("status, run_at, created_at, started_at, completed_at) ")
	query.WriteString("VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")

	res, err := tx.Exec(query.String(), p.ID, p.Name, p.Description, p.Status, p.RunAt, p.CreatedAt,
		p.StartedAt, p.CompletedAt)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		tx.Rollback()
		return fmt.Errorf("could not insert pipeline, rows affected: %d", rowsAffected)
	}

	for _, j := range p.Jobs {
		query.Reset()
		query.WriteString("INSERT INTO job (id, pipeline_id, next_job_id, name, task_name, task_params, ")
		query.WriteString("use_previous_results, timeout, description, status, failure_reason, run_at, ")
		query.WriteString("scheduled_at, created_at, started_at, completed_at) ")
		query.WriteString("VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)")

		var taskParams relational.MapStringInterface = j.TaskParams
		res, err := tx.Exec(query.String(), j.ID, j.PipelineID, j.NextJobID, j.Name, j.TaskName, taskParams,
			j.UsePreviousResults, j.Timeout, j.Description, j.Status, j.FailureReason, j.RunAt,
			j.ScheduledAt, j.CreatedAt, j.StartedAt, j.CompletedAt)
		if err != nil {
			tx.Rollback()
			return err
		}
		if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
			tx.Rollback()
			return fmt.Errorf("could not insert jobs, rows affected: %d", rowsAffected)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetPipeline fetches a pipeline from the storage.
func (psql *PostgreSQL) GetPipeline(id string) (*domain.Pipeline, error) {
	var query bytes.Buffer
	query.WriteString("SELECT id, name, description, status, run_at, ")
	query.WriteString("created_at, started_at, completed_at ")
	query.WriteString("FROM pipeline WHERE id=$1")

	p := new(domain.Pipeline)

	err := psql.DB.QueryRow(query.String(), id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Status, &p.RunAt,
		&p.CreatedAt, &p.StartedAt, &p.CompletedAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "pipeline"}
	}
	return p, nil
}

// GetPipelines fetches all pipelines from the storage, optionally filters the pipelines by status.
func (psql *PostgreSQL) GetPipelines(status domain.JobStatus) ([]*domain.Pipeline, error) {
	filterByStatus := ""
	if status != domain.Undefined {
		filterByStatus = fmt.Sprintf("WHERE status = %d", status.Index())
	}

	var query bytes.Buffer
	query.WriteString("SELECT id, name, description, status, run_at, ")
	query.WriteString("created_at, started_at, completed_at ")
	query.WriteString("FROM pipeline " + filterByStatus + " ORDER BY created_at ASC")

	pipelines := make([]*domain.Pipeline, 0)

	rows, err := psql.DB.Query(query.String())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		p := new(domain.Pipeline)
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.RunAt,
			&p.CreatedAt, &p.StartedAt, &p.CompletedAt)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, nil
}

// UpdatePipeline updates a pipeline to the storage.
func (psql *PostgreSQL) UpdatePipeline(id string, p *domain.Pipeline) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("UPDATE pipeline SET name=$1, description=$2, status=$3, ")
	query.WriteString("run_at=$4, created_at=$5, started_at=$6, completed_at=$7 ")
	query.WriteString("WHERE id=$8")

	res, err := tx.Exec(query.String(), p.Name, p.Description, p.Status,
		p.RunAt, p.CreatedAt, p.StartedAt, p.CompletedAt, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
		tx.Rollback()
		return fmt.Errorf("could not update pipeline, rows affected: %d", rowsAffected)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeletePipeline deletes a pipeline and all its jobs from the storage.
func (psql *PostgreSQL) DeletePipeline(id string) error {
	tx, err := psql.DB.Begin()
	if err != nil {
		return err
	}

	// CASCADE
	var query bytes.Buffer
	query.WriteString("SELECT id ")
	query.WriteString("FROM job WHERE pipeline_id=$1 ORDER BY created_at ASC")

	jobIDs := make([]string, 0)

	rows, err := psql.DB.Query(query.String(), id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	for rows.Next() {
		var jobID string
		err = rows.Scan(&jobID)
		if err != nil {
			return err
		}
		jobIDs = append(jobIDs, jobID)
	}

	for _, jobID := range jobIDs {
		query.Reset()
		query.WriteString("DELETE FROM jobresult WHERE job_id=$1")
		if _, err = tx.Exec(query.String(), jobID); err != nil {
			tx.Rollback()
			return err
		}
	}

	query.Reset()
	query.WriteString("DELETE FROM job WHERE pipeline_id=$1")
	if _, err = tx.Exec(query.String(), id); err != nil {
		tx.Rollback()
		return err
	}

	query.Reset()
	query.WriteString("DELETE FROM pipeline WHERE id=$1")
	if _, err = tx.Exec(query.String(), id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func createDB(parsedDSN, dbName string) {
	dsnWithoutDBName := strings.Replace(parsedDSN, dbName, "", 1)

	db, err := sql.Open("postgres", dsnWithoutDBName)
	if err != nil {
		panic(err)
	}

	var query bytes.Buffer
	query.WriteString(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	_, err = db.Exec(query.String())
	if err != nil {
		panic(err)
	}

	query.Reset()
	query.WriteString(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	_, err = db.Exec(query.String())
	if err != nil {
		panic(err)
	}
	db.Close()

	db, err = sql.Open("postgres", parsedDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	// Create pipeline table
	query.Reset()
	query.WriteString(createPipelineTableMigration)
	_, err = tx.Exec(query.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// Create job table
	query.Reset()
	query.WriteString(createJobTableMigration)
	_, err = tx.Exec(query.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// Create jobresult table
	query.Reset()
	query.WriteString(createJobResultTableMigration)
	_, err = tx.Exec(query.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// Create indexes
	query.Reset()
	query.WriteString(createIndexes)
	_, err = tx.Exec(query.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
}
