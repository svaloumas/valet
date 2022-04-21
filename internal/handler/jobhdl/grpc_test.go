package jobhdl

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/svaloumas/valet/internal/core/domain"
	pb "github.com/svaloumas/valet/internal/handler/jobhdl/protobuf"
	"github.com/svaloumas/valet/mock"
	"github.com/svaloumas/valet/pkg/apperrors"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGRPCCreateJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)
	jobService := mock.NewMockJobService(ctrl)

	handler := &JobgRPCHandler{
		jobService: jobService,
	}
	pb.RegisterJobServer(s, handler)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "testnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("failed to dial testnet: %v", err)
	}
	defer conn.Close()

	taskParams := map[string]interface{}{
		"url": "some-url.com",
	}
	runAt, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z")
	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams:  taskParams,
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	jobWithSchedule := &domain.Job{}
	*jobWithSchedule = *job
	jobWithSchedule.RunAt = &runAt

	client := pb.NewJobClient(conn)
	taskParamsStruct, err := structpb.NewStruct(taskParams)
	if err != nil {
		t.Errorf("failed to create task params struct: %v", err)
	}

	jobServiceErr := errors.New("some job service error")
	jobValidationErr := &apperrors.ResourceValidationErr{Message: "some job validation error"}
	fullQueueErr := &apperrors.FullQueueErr{}
	parseTimeErr := &apperrors.ParseTimeErr{}

	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(job, nil).
		Times(1)
	jobService.
		EXPECT().
		Create(jobWithSchedule.Name, jobWithSchedule.TaskName, jobWithSchedule.Description, "2006-01-02T15:04:05.999999999Z", jobWithSchedule.Timeout, taskParams).
		Return(jobWithSchedule, nil).
		Times(1)
	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(nil, jobServiceErr).
		Times(1)
	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(nil, fullQueueErr).
		Times(1)
	jobService.
		EXPECT().
		Create("", job.TaskName, job.Description, "", job.Timeout, job.TaskParams).
		Return(nil, jobValidationErr).
		Times(1)
	jobService.
		EXPECT().
		Create(job.Name, job.TaskName, job.Description, "invalid_timestamp_format", job.Timeout, job.TaskParams).
		Return(nil, parseTimeErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.CreateJobRequest
		res  *pb.CreateJobResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.CreateJobRequest{
				Name:        job.Name,
				Description: job.Description,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
			},
			&pb.CreateJobResponse{
				Id:          job.ID,
				Name:        job.Name,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
				Description: job.Description,
				Status:      int32(job.Status),
				CreatedAt:   timestamppb.New(*job.CreatedAt),
			},
			codes.OK,
			nil,
		},
		{
			"ok with schedule",
			&pb.CreateJobRequest{
				Name:        jobWithSchedule.Name,
				Description: jobWithSchedule.Description,
				TaskName:    jobWithSchedule.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
				RunAt:       "2006-01-02T15:04:05.999999999Z",
			},
			&pb.CreateJobResponse{
				Id:          jobWithSchedule.ID,
				Name:        jobWithSchedule.Name,
				TaskName:    jobWithSchedule.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(jobWithSchedule.Timeout),
				Description: jobWithSchedule.Description,
				Status:      int32(jobWithSchedule.Status),
				RunAt:       timestamppb.New(*jobWithSchedule.RunAt),
				CreatedAt:   timestamppb.New(*jobWithSchedule.CreatedAt),
			},
			codes.OK,
			nil,
		},
		{
			"internal error",
			&pb.CreateJobRequest{
				Name:        job.Name,
				Description: job.Description,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
			},
			nil,
			codes.Internal,
			status.Error(codes.Internal, jobServiceErr.Error()),
		},
		{
			"job validation error",
			&pb.CreateJobRequest{
				Description: job.Description,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
			},
			nil,
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, jobValidationErr.Error()),
		},
		{
			"service unavailable error",
			&pb.CreateJobRequest{
				Name:        job.Name,
				Description: job.Description,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
			},
			nil,
			codes.Unavailable,
			status.Error(codes.Unavailable, fullQueueErr.Error()),
		},
		{
			"invalid timestamp format",
			&pb.CreateJobRequest{
				Name:        job.Name,
				Description: job.Description,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
				RunAt:       "invalid_timestamp_format",
			},
			nil,
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, parseTimeErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Create(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("job create returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("job create returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("job create returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCGetJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)
	jobService := mock.NewMockJobService(ctrl)

	handler := &JobgRPCHandler{
		jobService: jobService,
	}
	pb.RegisterJobServer(s, handler)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "testnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("failed to dial testnet: %v", err)
	}
	defer conn.Close()

	taskParams := map[string]interface{}{
		"url": "some-url.com",
	}
	runAt, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z")
	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams:  taskParams,
		Status:      domain.Pending,
		RunAt:       &runAt,
		CreatedAt:   &createdAt,
	}

	client := pb.NewJobClient(conn)

	taskParamsStruct, err := structpb.NewStruct(taskParams)
	if err != nil {
		t.Errorf("failed to create task params struct: %v", err)
	}

	jobNotFoundErr := &apperrors.NotFoundErr{ID: job.ID, ResourceName: "job"}
	jobServiceErr := errors.New("some job service error")

	jobService.
		EXPECT().
		Get(job.ID).
		Return(job, nil).
		Times(1)
	jobService.
		EXPECT().
		Get("invalid_id").
		Return(nil, jobNotFoundErr).
		Times(1)
	jobService.
		EXPECT().
		Get(job.ID).
		Return(nil, jobServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.GetJobRequest
		res  *pb.GetJobResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.GetJobRequest{Id: job.ID},
			&pb.GetJobResponse{
				Id:          job.ID,
				Name:        job.Name,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
				Description: job.Description,
				Status:      int32(job.Status),
				RunAt:       timestamppb.New(*job.RunAt),
				CreatedAt:   timestamppb.New(*job.CreatedAt),
			},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.GetJobRequest{Id: "invalid_id"},
			nil,
			codes.NotFound,
			status.Error(codes.NotFound, jobNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.GetJobRequest{Id: job.ID},
			nil,
			codes.Internal,
			status.Error(codes.Internal, jobServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Get(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("job get returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("job get returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("job get returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCGetJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(1)
	jobService := mock.NewMockJobService(ctrl)

	handler := &JobgRPCHandler{
		jobService: jobService,
	}
	pb.RegisterJobServer(s, handler)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "testnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("failed to dial testnet: %v", err)
	}
	defer conn.Close()

	taskParams := map[string]interface{}{
		"url": "some-url.com",
	}
	runAt, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z")
	createdAt := freezed.Now()
	pendingJob := &domain.Job{
		ID:          "auuid4",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams:  taskParams,
		Status:      domain.Pending,
		RunAt:       &runAt,
		CreatedAt:   &createdAt,
	}

	pendingJobs := []*domain.Job{pendingJob}
	client := pb.NewJobClient(conn)

	taskParamsStruct, err := structpb.NewStruct(taskParams)
	if err != nil {
		t.Errorf("failed to create task params struct: %v", err)
	}

	jobStatusValidationErr := &apperrors.ResourceValidationErr{Message: "invalid job status: \"NOT_STARTED\""}
	jobServiceErr := errors.New("some job service error")

	jobService.
		EXPECT().
		GetJobs("pending").
		Return(pendingJobs, nil).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("not_started").
		Return(nil, jobStatusValidationErr).
		Times(1)
	jobService.
		EXPECT().
		GetJobs("pending").
		Return(nil, jobServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.GetJobsRequest
		res  *pb.GetJobsResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.GetJobsRequest{Status: "pending"},
			&pb.GetJobsResponse{
				Jobs: []*pb.GetJobResponse{
					&pb.GetJobResponse{
						Id:          pendingJob.ID,
						Name:        pendingJob.Name,
						TaskName:    pendingJob.TaskName,
						TaskParams:  taskParamsStruct,
						Timeout:     int32(pendingJob.Timeout),
						Description: pendingJob.Description,
						Status:      int32(pendingJob.Status),
						RunAt:       timestamppb.New(*pendingJob.RunAt),
						CreatedAt:   timestamppb.New(*pendingJob.CreatedAt),
					},
				},
			},
			codes.OK,
			nil,
		},
		{
			"job status validation error",
			&pb.GetJobsRequest{Status: "not_started"},
			nil,
			codes.NotFound,
			status.Error(codes.InvalidArgument, jobStatusValidationErr.Error()),
		},
		{
			"internal error",
			&pb.GetJobsRequest{Status: "pending"},
			nil,
			codes.Internal,
			status.Error(codes.Internal, jobServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.GetJobs(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("get jobs returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("get jobs returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("get jobs returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCUpdateJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	jobService := mock.NewMockJobService(ctrl)

	handler := &JobgRPCHandler{
		jobService: jobService,
	}
	pb.RegisterJobServer(s, handler)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "testnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("failed to dial testnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewJobClient(conn)

	jobID := "auuid4"
	jobNotFoundErr := &apperrors.NotFoundErr{ID: jobID, ResourceName: "job"}
	jobServiceErr := errors.New("some job service error")

	jobService.
		EXPECT().
		Update(jobID, "updated_name", "updated description").
		Return(nil).
		Times(1)
	jobService.
		EXPECT().
		Update("invalid_id", "updated_name", "updated description").
		Return(jobNotFoundErr).
		Times(1)
	jobService.
		EXPECT().
		Update(jobID, "updated_name", "updated description").
		Return(jobServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.UpdateJobRequest
		res  *pb.UpdateJobResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.UpdateJobRequest{
				Id:          jobID,
				Name:        "updated_name",
				Description: "updated description",
			},
			&pb.UpdateJobResponse{},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.UpdateJobRequest{
				Id:          "invalid_id",
				Name:        "updated_name",
				Description: "updated description",
			},
			&pb.UpdateJobResponse{},
			codes.NotFound,
			status.Error(codes.NotFound, jobNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.UpdateJobRequest{
				Id:          jobID,
				Name:        "updated_name",
				Description: "updated description",
			},
			&pb.UpdateJobResponse{},
			codes.Internal,
			status.Error(codes.Internal, jobServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Update(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("job update returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("job update returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("job update returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCDeleteJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	jobService := mock.NewMockJobService(ctrl)

	handler := &JobgRPCHandler{
		jobService: jobService,
	}
	pb.RegisterJobServer(s, handler)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "testnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("failed to dial testnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewJobClient(conn)

	jobID := "auuid4"
	jobNotFoundErr := &apperrors.NotFoundErr{ID: jobID, ResourceName: "job"}
	jobServiceErr := errors.New("some job service error")
	cannotDeletePipelineJobErr := &apperrors.CannotDeletePipelineJobErr{
		Message: "job with ID: pipeline_job_id can not be deleted because it belongs to a pipeline - try to delete the pipeline instead",
	}

	jobService.
		EXPECT().
		Delete(jobID).
		Return(nil).
		Times(1)
	jobService.
		EXPECT().
		Delete("invalid_id").
		Return(jobNotFoundErr).
		Times(1)
	jobService.
		EXPECT().
		Delete(jobID).
		Return(jobServiceErr).
		Times(1)
	jobService.
		EXPECT().
		Delete("pipeline_job_id").
		Return(cannotDeletePipelineJobErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.DeleteJobRequest
		res  *pb.DeleteJobResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.DeleteJobRequest{Id: jobID},
			&pb.DeleteJobResponse{},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.DeleteJobRequest{Id: "invalid_id"},
			nil,
			codes.NotFound,
			status.Error(codes.NotFound, jobNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.DeleteJobRequest{Id: jobID},
			nil,
			codes.Internal,
			status.Error(codes.Internal, jobServiceErr.Error()),
		},
		{
			"cannot delete pipeline job error",
			&pb.DeleteJobRequest{Id: "pipeline_job_id"},
			nil,
			codes.PermissionDenied,
			status.Error(codes.PermissionDenied, cannotDeletePipelineJobErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Delete(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("job delete returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("job delete returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("job delete returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}
