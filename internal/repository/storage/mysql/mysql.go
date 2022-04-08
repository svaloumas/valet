package mysql

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"
	"valet/internal/core/domain"
	"valet/pkg/apperrors"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// MySQLOptions represents config parameters for the MySQL client.
type MySQLOptions struct {
	ConnectionMaxLifetime int
	MaxIdleConnections    int
	MaxOpenConnections    int
}

// MySQL represents a MySQL DB client.
type MySQL struct {
	DB        *sql.DB
	Config    *mysql.Config
	Log       *logrus.Logger
	CaPemFile string
}

// New initializes and returns a MySQL client.
func New(
	dsn,
	caPemFile,
	migrationsPath string,
	options *MySQLOptions,
	logger *logrus.Logger) *MySQL {

	mySQL := new(MySQL)

	if logger == nil {
		mySQL.Log = &logrus.Logger{Out: ioutil.Discard}
	} else {
		mySQL.Log = logger
	}

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

	mySQL.DB, err = sql.Open("mysql", config.FormatDSN())
	if err != nil {
		panic(err)
	}
	mySQL.DB.SetConnMaxLifetime(time.Duration(options.ConnectionMaxLifetime) * time.Millisecond)
	mySQL.DB.SetMaxIdleConns(options.MaxIdleConnections)
	mySQL.DB.SetMaxOpenConns(options.MaxOpenConnections)

	err = mySQL.CreateDBIfNotExists()
	if err != nil {
		panic(err)
	}

	err = mySQL.Migrate(migrationsPath)
	if err != nil {
		panic(err)
	}

	return mySQL
}

// CheckHealth returns the status of MySQL.
func (mySQL *MySQL) CheckHealth() bool {

	err := mySQL.DB.Ping()
	return err == nil
}

// Close terminates any store connections gracefully.
func (mySQL *MySQL) Close() error {
	return mySQL.DB.Close()
}

// CreateJob adds new job to the repository.
func (storage *MySQL) CreateJob(j *domain.Job) error {
	tx, err := storage.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO job (id, name, task_name, task_params, ")
	query.WriteString("timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at) ")
	query.WriteString("VALUES (UuidToBin(?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

	var taskParams MapStringInterface = j.TaskParams
	if err != nil {
		return err
	}
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
func (storage *MySQL) GetJob(id string) (*domain.Job, error) {

	var query bytes.Buffer
	query.WriteString("SELECT UuidFromBin(id), name, task_name, task_params, ")
	query.WriteString("timeout, description, status, failure_reason, run_at, ")
	query.WriteString("scheduled_at, created_at, started_at, completed_at ")
	query.WriteString("FROM job WHERE id=UuidToBin(?)")

	var taskParams MapStringInterface
	job := new(domain.Job)

	err := storage.DB.QueryRow(query.String(), id).Scan(
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

// UpdateJob updates a job to the repository.
func (storage *MySQL) UpdateJob(id string, j *domain.Job) error {
	tx, err := storage.DB.Begin()
	if err != nil {
		return err
	}
	var query bytes.Buffer
	var taskParams MapStringInterface = j.TaskParams

	query.WriteString("UPDATE job SET name=?, task_name=?, task_params=?, timeout=?, ")
	query.WriteString("description=?, status=?, failure_reason=?, run_at=?, ")
	query.WriteString("scheduled_at=?, created_at=?, started_at=?, completed_at=? ")
	query.WriteString("WHERE id=UuidToBin(?)")

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
func (storage *MySQL) DeleteJob(id string) error {
	tx, err := storage.DB.Begin()
	if err != nil {
		return err
	}

	var query bytes.Buffer
	// query.WriteString("DELETE FROM jobresult WHERE job_id=UuidToBin(?)")
	// if _, err = tx.Exec(query.String(), id); err != nil {
	// 	tx.Rollback()
	// 	return err
	// }

	// query.Reset()
	query.WriteString("DELETE FROM job WHERE id=UuidToBin(?)")
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
func (storage *MySQL) GetDueJobs() ([]*domain.Job, error) {
	return nil, nil
}

// CreateJobResult adds new job result to the repository.
func (storage *MySQL) CreateJobResult(result *domain.JobResult) error {
	return nil
}

// GetJobResult fetches a job result from the repository.
func (storage *MySQL) GetJobResult(id string) (*domain.JobResult, error) {
	return nil, nil
}

// UpdateJobResult updates a job result to the repository.
func (storage *MySQL) UpdateJobResult(id string, result *domain.JobResult) error {
	return nil
}

// DeleteJobResult deletes a job result from the repository.
func (storage *MySQL) DeleteJobResult(id string) error {
	return nil
}
