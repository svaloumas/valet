package mysql

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/internal/repository/storage/relational"
	"valet/pkg/apperrors"
)

const (
	createJobTableMigration = `
		CREATE TABLE IF NOT EXISTS job (
		id binary(16) NOT NULL,
		name varchar(255) NOT NULL,
		task_name varchar(255) NOT NULL,
		task_params JSON NOT NULL,
		timeout INT NOT NULL,
		description varchar(255) NOT NULL DEFAULT '',
		status INT NOT NULL DEFAULT 0,
		failure_reason TEXT NOT NULL,
		run_at timestamp NULL ,
		scheduled_at timestamp NULL,
		created_at timestamp NOT NULL DEFAULT current_timestamp(),
		started_at timestamp NULL,
		completed_at timestamp NULL,
		PRIMARY KEY (` + "`id`" + `)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
	`
	createJobResultTableMigration = `
		CREATE TABLE IF NOT EXISTS jobresult (
		job_id binary(16) NOT NULL,
		metadata JSON NOT NULL,
		error TEXT NOT NULL,
		PRIMARY KEY (` + "`job_id`" + `),
		CONSTRAINT ` + "`fk_job_id` FOREIGN KEY (`job_id`) REFERENCES `job` (`id`) " + `
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
	`
)

var _ port.Storage = &MySQL{}

// MySQL represents a MySQL DB client.
type MySQL struct {
	DB        *sql.DB
	Config    *mysql.Config
	CaPemFile string
}

// New initializes and returns a MySQL client.
func New(dsn, caPemFile string, options *relational.DBOptions) *MySQL {
	mySQL := new(MySQL)

	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic(err)
	}
	mySQL.Config = config

	if caPemFile != "" {
		rootCertPool := x509.NewCertPool()
		pem, _ := ioutil.ReadFile(caPemFile)
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			panic(fmt.Errorf("failed to append PEM: %s", caPemFile))
		}
		mysql.RegisterTLSConfig("custom", &tls.Config{RootCAs: rootCertPool})
		mySQL.Config.TLSConfig = "custom"
		mySQL.CaPemFile = caPemFile
	}

	createDB(config.FormatDSN(), config.DBName)

	mySQL.DB, err = sql.Open("mysql", config.FormatDSN())
	if err != nil {
		panic(err)
	}
	mySQL.DB.SetConnMaxLifetime(time.Duration(options.ConnectionMaxLifetime) * time.Millisecond)
	mySQL.DB.SetMaxIdleConns(options.MaxIdleConnections)
	mySQL.DB.SetMaxOpenConns(options.MaxOpenConnections)
	return mySQL
}

// CheckHealth returns the status of MySQL.
func (mySQL *MySQL) CheckHealth() bool {
	err := mySQL.DB.Ping()
	return err == nil
}

// Close terminates any storage connections gracefully.
func (mySQL *MySQL) Close() error {
	return mySQL.DB.Close()
}

// CreateJob adds a new job to the repository.
func (mySQL *MySQL) CreateJob(j *domain.Job) error {
	tx, err := mySQL.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO job (id, name, task_name, task_params, ")
	query.WriteString("timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at) ")
	query.WriteString("VALUES (UUID_TO_BIN(?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

	var taskParams relational.MapStringInterface = j.TaskParams
	res, err := tx.Exec(query.String(), j.ID, j.Name, j.TaskName, taskParams,
		j.Timeout, j.Description, j.Status, j.FailureReason, j.RunAt,
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

// GetJob fetches a job from the repository.
func (mySQL *MySQL) GetJob(id string) (*domain.Job, error) {
	var query bytes.Buffer
	query.WriteString("SELECT BIN_TO_UUID(id), name, task_name, task_params, ")
	query.WriteString("timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job WHERE id=UUID_TO_BIN(?)")

	var taskParams relational.MapStringInterface
	job := new(domain.Job)

	err := mySQL.DB.QueryRow(query.String(), id).Scan(
		&job.ID, &job.Name, &job.TaskName, &taskParams, &job.Timeout,
		&job.Description, &job.Status, &job.FailureReason, &job.RunAt,
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

// GetJobs fetches all jobs from the repository, optionally filters the jobs by status.
func (mySQL *MySQL) GetJobs(status domain.JobStatus) ([]*domain.Job, error) {

	filterByStatus := ""
	if status != domain.Undefined {
		filterByStatus = fmt.Sprintf("WHERE status = %d", status.Index())
	}

	var query bytes.Buffer
	query.WriteString("SELECT BIN_TO_UUID(id), name, task_name, task_params, ")
	query.WriteString("timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job " + filterByStatus + " ORDER BY created_at ASC")

	jobs := make([]*domain.Job, 0)

	rows, err := mySQL.DB.Query(query.String())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var taskParams relational.MapStringInterface
		job := new(domain.Job)
		err = rows.Scan(&job.ID, &job.Name, &job.TaskName, &taskParams, &job.Timeout,
			&job.Description, &job.Status, &job.FailureReason, &job.RunAt,
			&job.ScheduledAt, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// UpdateJob updates a job to the repository.
func (mySQL *MySQL) UpdateJob(id string, j *domain.Job) error {
	tx, err := mySQL.DB.Begin()
	if err != nil {
		return err
	}
	var query bytes.Buffer
	var taskParams relational.MapStringInterface = j.TaskParams

	query.WriteString("UPDATE job SET name=?, task_name=?, task_params=?, timeout=?, ")
	query.WriteString("description=?, status=?, failure_reason=?, run_at=?, ")
	query.WriteString("scheduled_at=?, created_at=?, started_at=?, completed_at=? ")
	query.WriteString("WHERE id=UUID_TO_BIN(?)")

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

// DeleteJob deletes a job from the repository.
func (mySQL *MySQL) DeleteJob(id string) error {
	tx, err := mySQL.DB.Begin()
	if err != nil {
		return err
	}

	// CASCADE
	var query bytes.Buffer
	query.WriteString("DELETE FROM jobresult WHERE job_id=UUID_TO_BIN(?)")
	if _, err = tx.Exec(query.String(), id); err != nil {
		tx.Rollback()
		return err
	}

	query.Reset()
	query.WriteString("DELETE FROM job WHERE id=UUID_TO_BIN(?)")
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
func (mySQL *MySQL) GetDueJobs() ([]*domain.Job, error) {

	var query bytes.Buffer
	query.WriteString("SELECT BIN_TO_UUID(id), name, task_name, task_params, ")
	query.WriteString("timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job WHERE run_at IS NOT NULL AND run_at < ? ")
	query.WriteString("AND status = 1 ORDER BY run_at ASC")

	dueJobs := make([]*domain.Job, 0)

	rows, err := mySQL.DB.Query(query.String(), time.Now())
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var taskParams relational.MapStringInterface
		job := new(domain.Job)

		err := rows.Scan(
			&job.ID, &job.Name, &job.TaskName, &taskParams, &job.Timeout,
			&job.Description, &job.Status, &job.FailureReason, &job.RunAt,
			&job.ScheduledAt, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
		if err != nil {
			return nil, err
		}
		job.TaskParams = taskParams
		dueJobs = append(dueJobs, job)
	}

	return dueJobs, nil
}

// CreateJobResult adds new job result to the repository.
func (mySQL *MySQL) CreateJobResult(result *domain.JobResult) error {
	tx, err := mySQL.DB.Begin()
	if err != nil {
		return err
	}

	metadataBytes, err := json.Marshal(result.Metadata)
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO jobresult (job_id, metadata, error) ")
	query.WriteString("VALUES (UUID_TO_BIN(?), ?, ?)")

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

// GetJobResult fetches a job result from the repository.
func (mySQL *MySQL) GetJobResult(jobID string) (*domain.JobResult, error) {
	var query bytes.Buffer
	query.WriteString("SELECT BIN_TO_UUID(job_id), metadata, error ")
	query.WriteString("FROM jobresult WHERE job_id=UUID_TO_BIN(?)")

	var metadataBytes []byte
	result := new(domain.JobResult)

	err := mySQL.DB.QueryRow(query.String(), jobID).Scan(
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

// UpdateJobResult updates a job result to the repository.
func (mySQL *MySQL) UpdateJobResult(jobID string, result *domain.JobResult) error {
	tx, err := mySQL.DB.Begin()
	if err != nil {
		return err
	}
	metadataBytes, err := json.Marshal(result.Metadata)
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("UPDATE jobresult SET metadata=?, error=? ")
	query.WriteString("WHERE job_id=UUID_TO_BIN(?)")

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

// DeleteJobResult deletes a job result from the repository.
func (mySQL *MySQL) DeleteJobResult(jobID string) error {
	tx, err := mySQL.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("DELETE FROM jobresult WHERE job_id=UUID_TO_BIN(?)")
	if _, err = tx.Exec(query.String(), jobID); err != nil {
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

	db, err := sql.Open("mysql", dsnWithoutDBName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var sql bytes.Buffer
	sql.WriteString(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	_, err = db.Exec(sql.String())
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	sql.Reset()
	sql.WriteString(fmt.Sprintf("USE %s", dbName))
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// Create job table
	sql.Reset()
	sql.WriteString(createJobTableMigration)
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// Create jobresult table
	sql.Reset()
	sql.WriteString(createJobResultTableMigration)
	_, err = tx.Exec(sql.String())
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
}
