package pipelinehdl

import (
	"context"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/internal/handler/jobhdl"
	jobpb "github.com/svaloumas/valet/internal/handler/jobhdl/protobuf"
	pb "github.com/svaloumas/valet/internal/handler/pipelinehdl/protobuf"
	"github.com/svaloumas/valet/pkg/apperrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PipelinegRPCHandler is a gRPC handler that exposes job endpoints.
type PipelinegRPCHandler struct {
	pb.UnimplementedPipelineServer
	pipelineService port.PipelineService
	jobQueue        port.JobQueue
}

// NewPipelinegRPCHandler creates and returns a new PipelinegRPCHandler.
func NewPipelinegRPCHandler(pipelineService port.PipelineService, jobQueue port.JobQueue) *PipelinegRPCHandler {
	return &PipelinegRPCHandler{
		pipelineService: pipelineService,
		jobQueue:        jobQueue,
	}
}

// Create creates a new pipeline and all of its jobs.
func (hdl *PipelinegRPCHandler) Create(ctx context.Context, in *pb.CreatePipelineRequest) (*pb.CreatePipelineResponse, error) {
	name := in.GetName()
	description := in.GetDescription()
	runAt := in.GetRunAt()
	jobsRequests := in.GetJobs()

	jobs := make([]*domain.Job, 0)
	for _, req := range jobsRequests {
		job := &domain.Job{
			Name:        req.GetName(),
			Description: req.GetDescription(),
			TaskName:    req.GetTaskName(),
			TaskParams:  req.GetTaskParams().AsMap(),
			Timeout:     int(req.GetTimeout()),
		}
		jobs = append(jobs, job)
	}

	p, err := hdl.pipelineService.Create(name, description, runAt, jobs)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case *apperrors.ParseTimeErr:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if !p.IsScheduled() {
		// Push it as one job into the queue.
		p.MergeJobsInOne()

		// Push only the first job of the pipeline.
		if err := hdl.jobQueue.Push(p.Jobs[0]); err != nil {
			switch err.(type) {
			case *apperrors.FullQueueErr:
				return nil, status.Error(codes.Unavailable, err.Error())
			default:
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	}
	res, err := newCreatePipelineResponse(p)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}

// Get fetches a pipeline.
func (hdl *PipelinegRPCHandler) Get(ctx context.Context, in *pb.GetPipelineRequest) (*pb.GetPipelineResponse, error) {
	id := in.GetId()

	p, err := hdl.pipelineService.Get(id)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res, err := newGetPipelineResponse(p)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}

// GetPipelines fetches all pipelines, optionally filters them by status.
func (hdl *PipelinegRPCHandler) GetPipelines(ctx context.Context, in *pb.GetPipelinesRequest) (*pb.GetPipelinesResponse, error) {
	statusValue := in.GetStatus()

	pipelines, err := hdl.pipelineService.GetPipelines(statusValue)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res := &pb.GetPipelinesResponse{}
	for _, p := range pipelines {
		getPipelineResponse, err := newGetPipelineResponse(p)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		res.Pipelines = append(res.Pipelines, getPipelineResponse)
	}
	return res, nil
}

// GetPipelineJobs fetches the jobs of a specified pipeline.
func (hdl *PipelinegRPCHandler) GetPipelineJobs(ctx context.Context, in *pb.GetPipelineJobsRequest) (*pb.GetPipelineJobsResponse, error) {
	id := in.GetId()
	jobs, err := hdl.pipelineService.GetPipelineJobs(id)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res := &pb.GetPipelineJobsResponse{}
	for _, j := range jobs {
		getJobResponse, err := jobhdl.NewGetJobResponse(j)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		res.Jobs = append(res.Jobs, getJobResponse)
	}
	return res, nil
}

// Update updates a pipeline.
func (hdl *PipelinegRPCHandler) Update(ctx context.Context, in *pb.UpdatePipelineRequest) (*pb.UpdatePipelineResponse, error) {
	id := in.GetId()
	name := in.GetName()
	description := in.GetDescription()

	err := hdl.pipelineService.Update(id, name, description)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res := &pb.UpdatePipelineResponse{}
	return res, nil
}

// Delete deletes a pipelines and all its jobs.
func (hdl *PipelinegRPCHandler) Delete(ctx context.Context, in *pb.DeletePipelineRequest) (*pb.DeletePipelineResponse, error) {
	id := in.GetId()

	err := hdl.pipelineService.Delete(id)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res := &pb.DeletePipelineResponse{}
	return res, nil
}

func newCreatePipelineResponse(p *domain.Pipeline) (*pb.CreatePipelineResponse, error) {
	var runAtTimestamp *timestamppb.Timestamp
	var createdAtTimestamp *timestamppb.Timestamp

	if p.IsScheduled() {
		runAtTimestamp = timestamppb.New(*p.RunAt)
	}
	if p.CreatedAt != nil {
		createdAtTimestamp = timestamppb.New(*p.CreatedAt)
	}
	jobs := make([]*jobpb.CreateJobResponse, 0)
	for _, j := range p.Jobs {
		job, err := jobhdl.NewCreateJobResponse(j)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	res := &pb.CreatePipelineResponse{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Jobs:        jobs,
		Status:      int32(p.Status),
		RunAt:       runAtTimestamp,
		CreatedAt:   createdAtTimestamp,
	}
	return res, nil
}

func newGetPipelineResponse(p *domain.Pipeline) (*pb.GetPipelineResponse, error) {
	var runAtTimestamp *timestamppb.Timestamp
	var createdAtTimestamp *timestamppb.Timestamp
	var startedAtTimestamp *timestamppb.Timestamp
	var completedAtTimestamp *timestamppb.Timestamp
	var duration int64

	if p.IsScheduled() {
		runAtTimestamp = timestamppb.New(*p.RunAt)
	}
	if p.CreatedAt != nil {
		createdAtTimestamp = timestamppb.New(*p.CreatedAt)
	}
	if p.StartedAt != nil {
		startedAtTimestamp = timestamppb.New(*p.StartedAt)
	}
	if p.CompletedAt != nil {
		completedAtTimestamp = timestamppb.New(*p.CompletedAt)
	}
	if p.Duration != nil {
		duration = int64(*p.Duration)
	}
	res := &pb.GetPipelineResponse{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      int32(p.Status),
		RunAt:       runAtTimestamp,
		CreatedAt:   createdAtTimestamp,
		StartedAt:   startedAtTimestamp,
		CompletedAt: completedAtTimestamp,
		Duration:    duration,
	}
	return res, nil
}
