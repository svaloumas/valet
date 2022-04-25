package taskhdl

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/golang/mock/gomock"
	"github.com/svaloumas/valet/internal/core/service/tasksrv/taskrepo"
	pb "github.com/svaloumas/valet/internal/handler/taskhdl/protobuf"
	"github.com/svaloumas/valet/mock"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGRPCGetTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	taskService := mock.NewMockTaskService(ctrl)

	handler := NewTaskgRPCHandler(taskService)
	pb.RegisterTaskServer(s, handler)
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

	client := pb.NewTaskClient(conn)

	taskrepo := taskrepo.New()
	taskrepo.Register("test_task", func(...interface{}) (interface{}, error) {
		return "some metadata", nil
	})
	taskService.
		EXPECT().
		GetTaskRepository().
		Return(taskrepo).
		Times(1)

	req := &pb.GetTasksRequest{}
	response := &pb.GetTasksResponse{
		Tasks: []string{
			"test_task",
		},
	}

	res, err := client.GetTasks(ctx, req)
	if err != nil {
		t.Errorf("task get tasks returned unexpected error:%v", err)
	}

	serializedActualResponse, _ := json.Marshal(res)
	expected, _ := json.Marshal(response)
	if eq := reflect.DeepEqual(serializedActualResponse, expected); !eq {
		t.Errorf("job get returned wrong response: got %v want %#v", serializedActualResponse, expected)
	}
}
