package pipelinesrv

import (
	"time"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	"github.com/svaloumas/valet/pkg/apperrors"
	rtime "github.com/svaloumas/valet/pkg/time"
	"github.com/svaloumas/valet/pkg/uuidgen"
)

var _ port.PipelineService = &pipelineservice{}

type pipelineservice struct {
	storage  port.Storage
	taskrepo *taskrepo.TaskRepository
	uuidGen  uuidgen.UUIDGenerator
	time     rtime.Time
}

// New creates a new pipeline service.
func New(
	storage port.Storage,
	taskrepo *taskrepo.TaskRepository,
	uuidGen uuidgen.UUIDGenerator,
	time rtime.Time) *pipelineservice {
	return &pipelineservice{
		storage:  storage,
		taskrepo: taskrepo,
		uuidGen:  uuidGen,
		time:     time,
	}
}

// Create creates a new pipeline.
func (srv *pipelineservice) Create(name, description, runAt string, jobs []*domain.Job) (*domain.Pipeline, error) {
	pipelineUUID, err := srv.uuidGen.GenerateRandomUUIDString()
	if err != nil {
		return nil, err
	}

	jobIDs := make([]string, 0)
	for i := 0; i < len(jobs); i++ {
		jobUUID, err := srv.uuidGen.GenerateRandomUUIDString()
		if err != nil {
			return nil, err
		}
		jobIDs = append(jobIDs, jobUUID)
	}

	jobsToCreate := make([]*domain.Job, 0)
	for i, job := range jobs {
		var runAtTime time.Time

		// Propagate runAt only to first job.
		if runAt != "" && i == 0 {
			runAtTime, err = time.Parse(time.RFC3339Nano, runAt)
			if err != nil {
				return nil, &apperrors.ParseTimeErr{Message: err.Error()}
			}
		}

		jobID := jobIDs[i]
		nextJobID := ""
		if i < len(jobs)-1 {
			nextJobID = jobIDs[i+1]
		}

		createdAt := srv.time.Now()
		j := domain.NewJob(
			jobID, job.Name, job.TaskName, job.Description, pipelineUUID, nextJobID,
			job.Timeout, &runAtTime, &createdAt, job.UsePreviousResults, job.TaskParams)

		if err := j.Validate(srv.taskrepo); err != nil {
			return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
		}

		jobsToCreate = append(jobsToCreate, j)
	}

	createdAt := srv.time.Now()
	p := domain.NewPipeline(pipelineUUID, name, description, jobsToCreate, &createdAt)

	if err := p.Validate(); err != nil {
		return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
	}
	// Inherit first job's schedule timestamp.
	p.RunAt = jobsToCreate[0].RunAt

	if err := srv.storage.CreatePipeline(p); err != nil {
		return nil, err
	}
	return p, nil
}
