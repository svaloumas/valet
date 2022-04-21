package jobsrv

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	"github.com/svaloumas/valet/pkg/apperrors"
	rtime "github.com/svaloumas/valet/pkg/time"
	"github.com/svaloumas/valet/pkg/uuidgen"
)

var _ port.JobService = &jobservice{}

type jobservice struct {
	storage  port.Storage
	taskrepo *taskrepo.TaskRepository
	uuidGen  uuidgen.UUIDGenerator
	time     rtime.Time
}

// New creates a new job service.
func New(
	storage port.Storage,
	taskrepo *taskrepo.TaskRepository,
	uuidGen uuidgen.UUIDGenerator,
	time rtime.Time) *jobservice {
	return &jobservice{
		storage:  storage,
		taskrepo: taskrepo,
		uuidGen:  uuidGen,
		time:     time,
	}
}

// Create creates a new job.
func (srv *jobservice) Create(
	name, taskName, description, runAt string,
	timeout int, taskParams map[string]interface{}) (*domain.Job, error) {
	var runAtTime time.Time

	uuid, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}
	if runAt != "" {
		runAtTime, err = time.Parse(time.RFC3339Nano, runAt)
		if err != nil {
			return nil, &apperrors.ParseTimeErr{Message: err.Error()}
		}
	}
	createdAt := srv.time.Now()
	j := domain.NewJob(
		uuid, name, taskName, description,
		"", "", timeout, &runAtTime,
		&createdAt, false, taskParams)

	if err := j.Validate(srv.taskrepo); err != nil {
		return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
	}

	if err := srv.storage.CreateJob(j); err != nil {
		return nil, err
	}
	return j, nil
}

// Get fetches a job.
func (srv *jobservice) Get(id string) (*domain.Job, error) {
	j, err := srv.storage.GetJob(id)
	if err != nil {
		return nil, err
	}
	j.SetDuration()
	return j, nil
}

// GetJobs fetches all jobs, optionally filters the jobs by status.
func (srv *jobservice) GetJobs(status string) ([]*domain.Job, error) {
	var jobStatus domain.JobStatus
	if status == "" {
		jobStatus = domain.Undefined
	} else {
		err := json.Unmarshal([]byte("\""+strings.ToUpper(status)+"\""), &jobStatus)
		if err != nil {
			return nil, err
		}
	}
	jobs, err := srv.storage.GetJobs(jobStatus)
	if err != nil {
		return nil, err
	}
	for _, j := range jobs {
		j.SetDuration()
	}
	return jobs, nil
}

// Update updates a job.
func (srv *jobservice) Update(id, name, description string) error {
	j, err := srv.storage.GetJob(id)
	if err != nil {
		return err
	}
	j.Name = name
	j.Description = description
	return srv.storage.UpdateJob(id, j)
}

// Delete deletes a job.
func (srv *jobservice) Delete(id string) error {
	j, err := srv.storage.GetJob(id)
	if err != nil {
		return err
	}
	if j.BelongsToPipeline() {
		// TODO: Add a test case for this scenario.
		return &apperrors.CannotDeletePipelineJobErr{
			Message: fmt.Sprintf(
				`job with ID: %s can not be deleted because it belongs 
				to a pipeline - try to delete the pipeline instead`, id),
		}
	}
	return srv.storage.DeleteJob(id)
}
