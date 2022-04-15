package resulthdl

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"valet/internal/core/domain"
	"valet/internal/core/port"
	pb "valet/internal/handler/resulthdl/protobuf"
	"valet/pkg/apperrors"
)

// ResultgRPCHandler is a gRPC handler that exposes job result endpoints.
type ResultgRPCHandler struct {
	pb.UnimplementedJobResultServer
	resultService port.ResultService
}

// NewResultgRPCHandler creates and returns a new JobgRPCHandler.
func NewResultgRPCHandler(resultService port.ResultService) *ResultgRPCHandler {
	return &ResultgRPCHandler{
		resultService: resultService,
	}
}

// Get fetches a job result.
func (hdl *ResultgRPCHandler) Get(ctx context.Context, in *pb.GetJobResultRequest) (*pb.GetJobResultResponse, error) {
	jobID := in.GetJobId()
	j, err := hdl.resultService.Get(jobID)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res, err := newGetJobResultResponse(j)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}

// Delete deletes a job result.
func (hdl *ResultgRPCHandler) Delete(ctx context.Context, in *pb.DeleteJobResultRequest) (*pb.DeleteJobResultResponse, error) {
	jobID := in.GetJobId()

	err := hdl.resultService.Delete(jobID)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	res := &pb.DeleteJobResultResponse{}
	return res, nil
}

func newGetJobResultResponse(result *domain.JobResult) (*pb.GetJobResultResponse, error) {
	metadataValue, err := structpb.NewValue(result.Metadata)
	if err != nil {
		return nil, err
	}
	res := &pb.GetJobResultResponse{
		JobId:    result.JobID,
		Metadata: metadataValue,
		Error:    result.Error,
	}
	return res, nil
}
