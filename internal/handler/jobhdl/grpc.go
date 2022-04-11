package jobhdl

import (
	"context"
	"valet/internal/core/domain"
	"valet/internal/core/port"
	pb "valet/internal/handler/jobhdl/protobuf"
	"valet/pkg/apperrors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JobgRPCHandler is a gRPC handler that exposes job endpoints.
type JobgRPCHandler struct {
	pb.UnimplementedJobServer
	jobService port.JobService
}

// NewJobgRPCHandler creates and returns a new JobgRPCHandler.
func NewJobgRPCHandler(jobService port.JobService) *JobgRPCHandler {
	return &JobgRPCHandler{
		jobService: jobService,
	}
}

// // Create creates a new job.
func (hdl *JobgRPCHandler) Create(ctx context.Context, in *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	name := in.GetName()
	taskName := in.GetTaskName()
	description := in.GetDescription()
	runAt := in.GetRunAt()
	timeout := in.GetTimeout()
	taskParams := in.GetTaskParams()

	j, err := hdl.jobService.Create(
		name, taskName, description, runAt, int(timeout), taskParams.AsMap())
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case *apperrors.ParseTimeErr:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case *apperrors.FullQueueErr:
			return nil, status.Error(codes.Unavailable, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res, err := newCreateJobResponse(j)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}
func (hdl *JobgRPCHandler) Get(ctx context.Context, in *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	id := in.GetId()
	j, err := hdl.jobService.Get(id)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res, err := newGetJobResponse(j)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}

// Update updates a job.
func (hdl *JobgRPCHandler) Update(ctx context.Context, in *pb.UpdateJobRequest) (*pb.UpdateJobResponse, error) {
	id := in.GetId()
	name := in.GetName()
	description := in.GetDescription()

	err := hdl.jobService.Update(id, name, description)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res := &pb.UpdateJobResponse{}
	return res, nil
}

// Delete deletes a job.
func (hdl *JobgRPCHandler) Delete(ctx context.Context, in *pb.DeleteJobRequest) (*pb.DeleteJobResponse, error) {
	id := in.GetId()

	err := hdl.jobService.Delete(id)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res := &pb.DeleteJobResponse{}
	return res, nil
}

func newCreateJobResponse(j *domain.Job) (*pb.CreateJobResponse, error) {
	taskParamsStruct, err := structpb.NewStruct(j.TaskParams)
	if err != nil {
		return nil, err
	}
	var runAtTimestamp *timestamppb.Timestamp
	var createdAtTimestamp *timestamppb.Timestamp

	if j.RunAt != nil {
		runAtTimestamp = timestamppb.New(*j.RunAt)
	}
	if j.CreatedAt != nil {
		createdAtTimestamp = timestamppb.New(*j.CreatedAt)
	}
	res := &pb.CreateJobResponse{
		Id:          j.ID,
		Name:        j.Name,
		Description: j.Description,
		TaskName:    j.TaskName,
		Timeout:     int32(j.Timeout),
		TaskParams:  taskParamsStruct,
		Status:      int32(j.Status),
		RunAt:       runAtTimestamp,
		CreatedAt:   createdAtTimestamp,
	}
	return res, nil
}

func newGetJobResponse(j *domain.Job) (*pb.GetJobResponse, error) {
	taskParamsStruct, err := structpb.NewStruct(j.TaskParams)
	if err != nil {
		return nil, err
	}
	var runAtTimestamp *timestamppb.Timestamp
	var scheduledAtTimestamp *timestamppb.Timestamp
	var createdAtTimestamp *timestamppb.Timestamp
	var startedAtTimestamp *timestamppb.Timestamp
	var completedAtTimestamp *timestamppb.Timestamp
	var duration int64

	if j.RunAt != nil {
		runAtTimestamp = timestamppb.New(*j.RunAt)
	}
	if j.ScheduledAt != nil {
		scheduledAtTimestamp = timestamppb.New(*j.ScheduledAt)
	}
	if j.CreatedAt != nil {
		createdAtTimestamp = timestamppb.New(*j.CreatedAt)
	}
	if j.StartedAt != nil {
		startedAtTimestamp = timestamppb.New(*j.StartedAt)
	}
	if j.CompletedAt != nil {
		completedAtTimestamp = timestamppb.New(*j.CompletedAt)
	}
	if j.Duration != nil {
		duration = int64(*j.Duration)
	}
	res := &pb.GetJobResponse{
		Id:            j.ID,
		Name:          j.Name,
		Description:   j.Description,
		TaskName:      j.TaskName,
		Timeout:       int32(j.Timeout),
		TaskParams:    taskParamsStruct,
		Status:        int32(j.Status),
		FailureReason: j.FailureReason,
		RunAt:         runAtTimestamp,
		ScheduledAt:   scheduledAtTimestamp,
		CreatedAt:     createdAtTimestamp,
		StartedAt:     startedAtTimestamp,
		CompletedAt:   completedAtTimestamp,
		Duration:      duration,
	}
	return res, nil
}
