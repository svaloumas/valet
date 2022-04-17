package domain

import (
	"fmt"
	"strings"
	"time"
)

// Pipeline represents a sequence of async tasks.
type Pipeline struct {
	// ID is the auto-generated pipeline identifier in UUID4 format.
	ID string `json:"id"`

	// Name is the name of the pipeline.
	Name string `json:"name"`

	// Description gives some information about the pipeline.
	Description string `json:"description,omitempty"`

	// Jobs represents the jobs included to this pipeline.
	Jobs []*Job `json:"jobs"`

	// Status represents the status of the pipeline.
	// Jobs propagate their finite status to the pipeline.
	Status JobStatus `json:"status"`

	// RunAt is the UTC timestamp indicating the time for the pipeline to run.
	//  This property will be propagated to the first job of the pipeline.
	RunAt *time.Time `json:"run_at,omitempty"`

	// CreatedAt is the UTC timestamp of the pipeline creation.
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// StartedAt is the UTC timestamp of the moment the pipeline started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is the UTC timestamp of the moment the pipeline finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Duration indicates how much the pipeline took to complete.
	Duration *time.Duration `json:"duration,omitempty"`
}

func NewPipeline(id, name, description string, jobs []*Job, createdAt *time.Time) *Pipeline {
	p := &Pipeline{
		ID:          id,
		Name:        name,
		Description: description,
		Jobs:        jobs,
		Status:      Pending,
		CreatedAt:   createdAt,
	}
	return p
}

// MarkStarted updates the status and timestamp at the moment the pipeline started.
func (p *Pipeline) MarkStarted(startedAt *time.Time) {
	p.Status = InProgress
	p.StartedAt = startedAt
}

// MarkCompleted updates the status and timestamp at the moment the pipeline finished.
func (p *Pipeline) MarkCompleted(completedAt *time.Time) {
	p.Status = Completed
	p.CompletedAt = completedAt
}

// MarkFailed updates the status and timestamp at the moment the pipeline failed.
func (p *Pipeline) MarkFailed(failedAt *time.Time) {
	p.Status = Failed
	p.CompletedAt = failedAt
}

// SetDuration sets the duration of the pipeline if it's completed of failed.
func (p *Pipeline) SetDuration() {
	if p.Status == Completed || p.Status == Failed {
		duration := p.CompletedAt.Sub(*p.StartedAt) / time.Millisecond
		p.Duration = &duration
	}
}

// Validate perfoms basic sanity checks on the pipeline request payload.
func (p *Pipeline) Validate() error {
	var required []string

	if p.Name == "" {
		required = append(required, "name")
	}

	if len(p.Jobs) == 0 {
		required = append(required, "jobs")
	}

	if len(required) > 0 {
		return fmt.Errorf(strings.Join(required, ", ") + " required")
	}

	if len(p.Jobs) == 1 {
		return fmt.Errorf("pipeline shoud have at least 2 jobs, %d given", len(p.Jobs))
	}

	if p.Status != Undefined {
		err := p.Status.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) IsScheduled() bool {
	return p.RunAt != nil
}

func (p *Pipeline) CreateNestedJobPipeline() {
	for i := 0; i < len(p.Jobs)-1; i++ {
		p.Jobs[i].Next = p.Jobs[i+1]
	}
}
