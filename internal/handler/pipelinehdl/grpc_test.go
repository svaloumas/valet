package pipelinehdl

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
	jobpb "github.com/svaloumas/valet/internal/handler/jobhdl/protobuf"
	pb "github.com/svaloumas/valet/internal/handler/pipelinehdl/protobuf"
	"github.com/svaloumas/valet/mock"
	"github.com/svaloumas/valet/pkg/apperrors"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGRPCCreatePipeline(t *testing.T) {
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
	pipelineService := mock.NewMockPipelineService(ctrl)
	jobQueue := mock.NewMockJobQueue(ctrl)

	handler := &PipelinegRPCHandler{
		pipelineService: pipelineService,
		jobQueue:        jobQueue,
	}
	pb.RegisterPipelineServer(s, handler)
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
	createdAt := freezed.Now()
	job := &domain.Job{
		ID:          "job_id",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams:  taskParams,
		Status:      domain.Pending,
		CreatedAt:   &createdAt,
	}
	secondJob := &domain.Job{}
	*secondJob = *job
	secondJob.ID = "second_job_id"
	secondJob.Name = "second_job"

	job.Next = secondJob
	job.NextJobID = secondJob.ID

	jobRequest := &domain.Job{
		Name:        job.Name,
		Description: job.Description,
		TaskName:    job.TaskName,
		TaskParams:  job.TaskParams,
		Timeout:     job.Timeout,
	}
	secondJobRequest := &domain.Job{
		Name:        secondJob.Name,
		Description: secondJob.Description,
		TaskName:    secondJob.TaskName,
		TaskParams:  secondJob.TaskParams,
		Timeout:     secondJob.Timeout,
	}
	jobsRequest := []*domain.Job{
		jobRequest, secondJobRequest,
	}

	runAt, _ := time.Parse(time.RFC3339Nano, runAtTestTime)
	jobWithSchedule := &domain.Job{}
	*jobWithSchedule = *job
	jobWithSchedule.RunAt = &runAt
	jobWithSchedule.ID = secondJob.ID

	pipeline := &domain.Pipeline{
		ID:          "pipeline_id",
		Name:        "pipeline_name",
		Description: "some description",
		Jobs: []*domain.Job{
			job, secondJob,
		},
		Status:    domain.Pending,
		CreatedAt: &createdAt,
	}
	pipelineWithSchedule := &domain.Pipeline{}
	*pipelineWithSchedule = *pipeline
	pipelineWithSchedule.RunAt = &runAt
	pipelineWithSchedule.Jobs = []*domain.Job{
		jobWithSchedule, secondJob,
	}

	client := pb.NewPipelineClient(conn)
	taskParamsStruct, err := structpb.NewStruct(taskParams)
	if err != nil {
		t.Errorf("failed to create task params struct: %v", err)
	}

	pipelineServiceErr := errors.New("some pipeline service error")
	pipelineValidationErr := &apperrors.ResourceValidationErr{Message: "some pipeline validation error"}
	fullQueueErr := &apperrors.FullQueueErr{}
	parseTimeErr := &apperrors.ParseTimeErr{}
	jobQueueErr := errors.New("some job queue error")

	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "", jobsRequest).
		Return(pipeline, nil).
		Times(1)
	jobQueue.
		EXPECT().
		Push(job).
		Return(nil).
		Times(1)
	pipelineService.
		EXPECT().
		Create(pipelineWithSchedule.Name, pipelineWithSchedule.Description, runAtTestTime, jobsRequest).
		Return(pipelineWithSchedule, nil).
		Times(1)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "", jobsRequest).
		Return(nil, pipelineServiceErr).
		Times(1)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "", jobsRequest).
		Return(pipeline, nil).
		Times(1)
	jobQueue.
		EXPECT().
		Push(job).
		Return(fullQueueErr).
		Times(1)
	pipelineService.
		EXPECT().
		Create("", pipeline.Description, "", jobsRequest).
		Return(nil, pipelineValidationErr).
		Times(1)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "invalid_timestamp_format", jobsRequest).
		Return(nil, parseTimeErr).
		Times(1)
	pipelineService.
		EXPECT().
		Create(pipeline.Name, pipeline.Description, "", jobsRequest).
		Return(pipeline, nil).
		Times(1)
	jobQueue.
		EXPECT().
		Push(job).
		Return(jobQueueErr).
		Times(1)

	req := &pb.CreatePipelineRequest{
		Name:        pipeline.Name,
		Description: pipeline.Description,
		RunAt:       "",
		Jobs: []*jobpb.CreateJobRequest{
			{
				Name:        job.Name,
				Description: job.Description,
				TaskName:    job.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(job.Timeout),
			},
			{
				Name:        secondJob.Name,
				Description: secondJob.Description,
				TaskName:    secondJob.TaskName,
				TaskParams:  taskParamsStruct,
				Timeout:     int32(secondJob.Timeout),
			},
		},
	}
	reqWithSchedule := &pb.CreatePipelineRequest{}
	*reqWithSchedule = *req
	reqWithSchedule.RunAt = runAtTestTime

	reqNoPipelineName := &pb.CreatePipelineRequest{}
	*reqNoPipelineName = *req
	reqNoPipelineName.Name = ""

	reqInvalidRunAt := &pb.CreatePipelineRequest{}
	*reqInvalidRunAt = *req
	reqInvalidRunAt.RunAt = "invalid_timestamp_format"

	tests := []struct {
		name string
		req  *pb.CreatePipelineRequest
		res  *pb.CreatePipelineResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			req,
			&pb.CreatePipelineResponse{
				Id:          pipeline.ID,
				Name:        pipeline.Name,
				Description: pipeline.Description,
				Jobs: []*jobpb.CreateJobResponse{
					{
						Id:          job.ID,
						Name:        job.Name,
						Description: job.Description,
						TaskName:    job.TaskName,
						TaskParams:  taskParamsStruct,
						Timeout:     int32(job.Timeout),
						Status:      int32(job.Status),
						CreatedAt:   timestamppb.New(*job.CreatedAt),
					},
					{
						Id:          secondJob.ID,
						Name:        secondJob.Name,
						Description: secondJob.Description,
						TaskName:    secondJob.TaskName,
						TaskParams:  taskParamsStruct,
						Timeout:     int32(secondJob.Timeout),
						Status:      int32(secondJob.Status),
						CreatedAt:   timestamppb.New(*secondJob.CreatedAt),
					},
				},
				Status:    int32(pipeline.Status),
				CreatedAt: timestamppb.New(*pipeline.CreatedAt),
			},
			codes.OK,
			nil,
		},
		{
			"ok with schedule",
			reqWithSchedule,
			&pb.CreatePipelineResponse{
				Id:          pipeline.ID,
				Name:        pipeline.Name,
				Description: pipeline.Description,
				RunAt:       timestamppb.New(*jobWithSchedule.RunAt),
				Jobs: []*jobpb.CreateJobResponse{
					{
						Id:          jobWithSchedule.ID,
						Name:        jobWithSchedule.Name,
						Description: jobWithSchedule.Description,
						TaskName:    jobWithSchedule.TaskName,
						TaskParams:  taskParamsStruct,
						Timeout:     int32(jobWithSchedule.Timeout),
						Status:      int32(jobWithSchedule.Status),
						RunAt:       timestamppb.New(*jobWithSchedule.RunAt),
						CreatedAt:   timestamppb.New(*jobWithSchedule.CreatedAt),
					},
					{
						Id:          secondJob.ID,
						Name:        secondJob.Name,
						Description: secondJob.Description,
						TaskName:    secondJob.TaskName,
						TaskParams:  taskParamsStruct,
						Timeout:     int32(secondJob.Timeout),
						Status:      int32(secondJob.Status),
						CreatedAt:   timestamppb.New(*secondJob.CreatedAt),
					},
				},
				Status:    int32(pipeline.Status),
				CreatedAt: timestamppb.New(*pipeline.CreatedAt),
			},
			codes.OK,
			nil,
		},
		{
			"pipeline service internal error",
			req,
			nil,
			codes.Internal,
			status.Error(codes.Internal, pipelineServiceErr.Error()),
		},
		{
			"service unavailable error",
			req,
			nil,
			codes.Unavailable,
			status.Error(codes.Unavailable, fullQueueErr.Error()),
		},
		{
			"pipeline validation error",
			reqNoPipelineName,
			nil,
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, pipelineValidationErr.Error()),
		},
		{
			"invalid timestamp format",
			reqInvalidRunAt,
			nil,
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, parseTimeErr.Error()),
		},
		{
			"job queue internal error",
			req,
			nil,
			codes.Internal,
			status.Error(codes.Internal, jobQueueErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Create(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("pipeline create returned wrong status code: got %v want %v", status.Code(), tt.code)
						}
					}
					t.Errorf("pipeline create returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("pipeline create returned wrong response: got %v want %v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCGetPipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)
	pipelineService := mock.NewMockPipelineService(ctrl)

	handler := &PipelinegRPCHandler{
		pipelineService: pipelineService,
	}
	pb.RegisterPipelineServer(s, handler)
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

	runAt, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z")
	createdAt := freezed.Now()
	startedAt := freezed.Now()
	completedAt := startedAt.Add(1 * time.Minute)
	p := &domain.Pipeline{
		ID:          "auuid4",
		Name:        "job_name",
		Description: "some description",
		Status:      domain.Completed,
		RunAt:       &runAt,
		CreatedAt:   &createdAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}
	p.SetDuration()

	client := pb.NewPipelineClient(conn)

	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: p.ID, ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService.
		EXPECT().
		Get(p.ID).
		Return(p, nil).
		Times(1)
	pipelineService.
		EXPECT().
		Get("invalid_id").
		Return(nil, pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		Get(p.ID).
		Return(nil, pipelineServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.GetPipelineRequest
		res  *pb.GetPipelineResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.GetPipelineRequest{Id: p.ID},
			&pb.GetPipelineResponse{
				Id:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Status:      int32(p.Status),
				RunAt:       timestamppb.New(*p.RunAt),
				CreatedAt:   timestamppb.New(*p.CreatedAt),
				StartedAt:   timestamppb.New(*p.StartedAt),
				CompletedAt: timestamppb.New(*p.CompletedAt),
				Duration:    int64(*p.Duration),
			},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.GetPipelineRequest{Id: "invalid_id"},
			nil,
			codes.NotFound,
			status.Error(codes.NotFound, pipelineNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.GetPipelineRequest{Id: p.ID},
			nil,
			codes.Internal,
			status.Error(codes.Internal, pipelineServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Get(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("pipeline get returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("pipeline get returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("pipeline get returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCGetPipelines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)
	pipelineService := mock.NewMockPipelineService(ctrl)

	handler := &PipelinegRPCHandler{
		pipelineService: pipelineService,
	}
	pb.RegisterPipelineServer(s, handler)
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

	runAt, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z")
	createdAt := freezed.Now()
	startedAt := freezed.Now()
	completedAt := startedAt.Add(1 * time.Minute)
	completedPipeline := &domain.Pipeline{
		ID:          "auuid4",
		Name:        "job_name",
		Description: "some description",
		Status:      domain.Completed,
		RunAt:       &runAt,
		CreatedAt:   &createdAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}
	completedPipeline.SetDuration()

	completedPipelines := []*domain.Pipeline{completedPipeline}
	client := pb.NewPipelineClient(conn)

	pipelineStatusValidationErr := &apperrors.ResourceValidationErr{Message: "invalid status: \"NOT_STARTED\""}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService.
		EXPECT().
		GetPipelines("completed").
		Return(completedPipelines, nil).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelines("not_started").
		Return(nil, pipelineStatusValidationErr).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelines("pending").
		Return(nil, pipelineServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.GetPipelinesRequest
		res  *pb.GetPipelinesResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.GetPipelinesRequest{Status: "completed"},
			&pb.GetPipelinesResponse{
				Pipelines: []*pb.GetPipelineResponse{
					&pb.GetPipelineResponse{
						Id:          completedPipeline.ID,
						Name:        completedPipeline.Name,
						Description: completedPipeline.Description,
						Status:      int32(completedPipeline.Status),
						RunAt:       timestamppb.New(*completedPipeline.RunAt),
						CreatedAt:   timestamppb.New(*completedPipeline.CreatedAt),
						StartedAt:   timestamppb.New(*completedPipeline.StartedAt),
						CompletedAt: timestamppb.New(*completedPipeline.CompletedAt),
						Duration:    int64(*completedPipeline.Duration),
					},
				},
			},
			codes.OK,
			nil,
		},
		{
			"pipeline status validation error",
			&pb.GetPipelinesRequest{Status: "not_started"},
			nil,
			codes.InvalidArgument,
			status.Error(codes.InvalidArgument, pipelineStatusValidationErr.Error()),
		},
		{
			"internal error",
			&pb.GetPipelinesRequest{Status: "pending"},
			nil,
			codes.Internal,
			status.Error(codes.Internal, pipelineServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.GetPipelines(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("get pipelines returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("get pipelines returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("get pipelines returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCGetPipelineJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	freezed := mock.NewMockTime(ctrl)
	freezed.
		EXPECT().
		Now().
		Return(time.Date(1985, 05, 04, 04, 32, 53, 651387234, time.UTC)).
		Times(2)
	pipelineService := mock.NewMockPipelineService(ctrl)

	handler := &PipelinegRPCHandler{
		pipelineService: pipelineService,
	}
	pb.RegisterPipelineServer(s, handler)
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
	startedAt := freezed.Now()
	completedAt := startedAt.Add(1 * time.Minute)
	job := &domain.Job{
		ID:          "auuid4",
		PipelineID:  "pipeline_id",
		Name:        "job_name",
		TaskName:    "test_task",
		Timeout:     10,
		Description: "some description",
		TaskParams:  taskParams,
		Status:      domain.Completed,
		RunAt:       &runAt,
		CreatedAt:   &createdAt,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}
	job.SetDuration()

	pipelineJobs := []*domain.Job{job}
	client := pb.NewPipelineClient(conn)

	taskParamsStruct, err := structpb.NewStruct(taskParams)
	if err != nil {
		t.Errorf("failed to create task params struct: %v", err)
	}

	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: job.PipelineID, ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some job service error")

	pipelineService.
		EXPECT().
		GetPipelineJobs(job.PipelineID).
		Return(pipelineJobs, nil).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelineJobs("invalid_id").
		Return(nil, pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		GetPipelineJobs(job.PipelineID).
		Return(nil, pipelineServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.GetPipelineJobsRequest
		res  *pb.GetPipelineJobsResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.GetPipelineJobsRequest{Id: job.PipelineID},
			&pb.GetPipelineJobsResponse{
				Jobs: []*jobpb.GetJobResponse{
					&jobpb.GetJobResponse{
						Id:          job.ID,
						Name:        job.Name,
						TaskName:    job.TaskName,
						TaskParams:  taskParamsStruct,
						Timeout:     int32(job.Timeout),
						Description: job.Description,
						Status:      int32(job.Status),
						RunAt:       timestamppb.New(*job.RunAt),
						CreatedAt:   timestamppb.New(*job.CreatedAt),
						StartedAt:   timestamppb.New(*job.StartedAt),
						CompletedAt: timestamppb.New(*job.CompletedAt),
						Duration:    int64(*job.Duration),
					},
				},
			},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.GetPipelineJobsRequest{Id: "invalid_id"},
			nil,
			codes.NotFound,
			status.Error(codes.NotFound, pipelineNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.GetPipelineJobsRequest{Id: job.PipelineID},
			nil,
			codes.Internal,
			status.Error(codes.Internal, pipelineServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.GetPipelineJobs(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("get pipeline jobs returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("get pipeline jobs returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("get pipeline jobs returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCUpdatePipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	pipelineService := mock.NewMockPipelineService(ctrl)

	handler := &PipelinegRPCHandler{
		pipelineService: pipelineService,
	}
	pb.RegisterPipelineServer(s, handler)
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

	client := pb.NewPipelineClient(conn)

	pipelineID := "auuid4"
	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: pipelineID, ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService.
		EXPECT().
		Update(pipelineID, "updated_name", "updated description").
		Return(nil).
		Times(1)
	pipelineService.
		EXPECT().
		Update("invalid_id", "updated_name", "updated description").
		Return(pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		Update(pipelineID, "updated_name", "updated description").
		Return(pipelineServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.UpdatePipelineRequest
		res  *pb.UpdatePipelineResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.UpdatePipelineRequest{
				Id:          pipelineID,
				Name:        "updated_name",
				Description: "updated description",
			},
			&pb.UpdatePipelineResponse{},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.UpdatePipelineRequest{
				Id:          "invalid_id",
				Name:        "updated_name",
				Description: "updated description",
			},
			&pb.UpdatePipelineResponse{},
			codes.NotFound,
			status.Error(codes.NotFound, pipelineNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.UpdatePipelineRequest{
				Id:          pipelineID,
				Name:        "updated_name",
				Description: "updated description",
			},
			&pb.UpdatePipelineResponse{},
			codes.Internal,
			status.Error(codes.Internal, pipelineServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Update(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("pipeline update returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("pipeline update returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("pipeline update returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}

func TestGRPCDeletePipeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	pipelineService := mock.NewMockPipelineService(ctrl)

	handler := &PipelinegRPCHandler{
		pipelineService: pipelineService,
	}
	pb.RegisterPipelineServer(s, handler)
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

	client := pb.NewPipelineClient(conn)

	pipelineID := "auuid4"
	pipelineNotFoundErr := &apperrors.NotFoundErr{ID: pipelineID, ResourceName: "pipeline"}
	pipelineServiceErr := errors.New("some pipeline service error")

	pipelineService.
		EXPECT().
		Delete(pipelineID).
		Return(nil).
		Times(1)
	pipelineService.
		EXPECT().
		Delete("invalid_id").
		Return(pipelineNotFoundErr).
		Times(1)
	pipelineService.
		EXPECT().
		Delete(pipelineID).
		Return(pipelineServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.DeletePipelineRequest
		res  *pb.DeletePipelineResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.DeletePipelineRequest{Id: pipelineID},
			&pb.DeletePipelineResponse{},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.DeletePipelineRequest{Id: "invalid_id"},
			nil,
			codes.NotFound,
			status.Error(codes.NotFound, pipelineNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.DeletePipelineRequest{Id: pipelineID},
			nil,
			codes.Internal,
			status.Error(codes.Internal, pipelineServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Delete(ctx, tt.req)
			if err != nil {
				if err.Error() != tt.err.Error() {
					if status, ok := status.FromError(err); ok {
						if status.Code() != tt.code {
							t.Errorf("pipeline delete returned wrong status code: got %v want %#v", status.Code(), tt.code)
						}
					}
					t.Errorf("pipeline delete returned wrong error: got %v want %#v", err.Error(), tt.err.Error())
				}
			} else {
				expectedSerialized, _ := json.Marshal(tt.res)
				actualSerialized, _ := json.Marshal(res)
				if eq := reflect.DeepEqual(actualSerialized, expectedSerialized); !eq {
					t.Errorf("pipeline delete returned wrong response: got %v want %#v", res, tt.res)
				}
			}
		})
	}
}
