package resulthdl

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/svaloumas/valet/internal/core/domain"
	pb "github.com/svaloumas/valet/internal/handler/resulthdl/protobuf"
	"github.com/svaloumas/valet/mock"
	"github.com/svaloumas/valet/pkg/apperrors"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGRPCGetJobResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	resultService := mock.NewMockResultService(ctrl)

	handler := &ResultgRPCHandler{
		resultService: resultService,
	}
	pb.RegisterJobResultServer(s, handler)
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

	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}
	metadataValue, _ := structpb.NewValue(result.Metadata)

	client := pb.NewJobResultClient(conn)

	resultNotFoundErr := &apperrors.NotFoundErr{ID: result.JobID, ResourceName: "job result"}
	resultServiceErr := errors.New("some result service error")

	resultService.
		EXPECT().
		Get(result.JobID).
		Return(result, nil).
		Times(1)
	resultService.
		EXPECT().
		Get("invalid_id").
		Return(nil, resultNotFoundErr).
		Times(1)
	resultService.
		EXPECT().
		Get(result.JobID).
		Return(nil, resultServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.GetJobResultRequest
		res  *pb.GetJobResultResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.GetJobResultRequest{JobId: result.JobID},
			&pb.GetJobResultResponse{
				JobId:    result.JobID,
				Metadata: metadataValue,
				Error:    result.Error,
			},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.GetJobResultRequest{JobId: "invalid_id"},
			nil,
			codes.NotFound,
			status.Error(codes.NotFound, resultNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.GetJobResultRequest{JobId: result.JobID},
			nil,
			codes.Internal,
			status.Error(codes.Internal, resultServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Get(ctx, tt.req)
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

func TestGRPCDeleteJobResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	resultService := mock.NewMockResultService(ctrl)

	handler := &ResultgRPCHandler{
		resultService: resultService,
	}
	pb.RegisterJobResultServer(s, handler)
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

	result := &domain.JobResult{
		JobID:    "auuid4",
		Metadata: "some metadata",
		Error:    "some task error",
	}

	client := pb.NewJobResultClient(conn)

	resultNotFoundErr := &apperrors.NotFoundErr{ID: result.JobID, ResourceName: "job result"}
	resultServiceErr := errors.New("some result service error")

	resultService.
		EXPECT().
		Delete(result.JobID).
		Return(nil).
		Times(1)
	resultService.
		EXPECT().
		Delete("invalid_id").
		Return(resultNotFoundErr).
		Times(1)
	resultService.
		EXPECT().
		Delete(result.JobID).
		Return(resultServiceErr).
		Times(1)

	tests := []struct {
		name string
		req  *pb.DeleteJobResultRequest
		res  *pb.DeleteJobResultResponse
		code codes.Code
		err  error
	}{
		{
			"ok",
			&pb.DeleteJobResultRequest{JobId: result.JobID},
			&pb.DeleteJobResultResponse{},
			codes.OK,
			nil,
		},
		{
			"not found",
			&pb.DeleteJobResultRequest{JobId: "invalid_id"},
			nil,
			codes.NotFound,
			status.Error(codes.NotFound, resultNotFoundErr.Error()),
		},
		{
			"internal error",
			&pb.DeleteJobResultRequest{JobId: result.JobID},
			nil,
			codes.Internal,
			status.Error(codes.Internal, resultServiceErr.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Delete(ctx, tt.req)
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
