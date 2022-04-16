package taskhdl

import (
	"context"

	"github.com/svaloumas/valet/internal/core/port"
	pb "github.com/svaloumas/valet/internal/handler/taskhdl/protobuf"
)

// TaskgRPCHandler is a gRPC handler that exposes task endpoints.
type TaskgRPCHandler struct {
	pb.UnimplementedTaskServer
	taskService port.TaskService
}

// NewTaskgRPCHandler creates and returns a new TaskgRPCHandler.
func NewTaskgRPCHandler(taskService port.TaskService) *TaskgRPCHandler {
	return &TaskgRPCHandler{
		taskService: taskService,
	}
}

func (hdl *TaskgRPCHandler) GetTasks(ctx context.Context, in *pb.GetTasksRequest) (*pb.GetTasksResponse, error) {
	taskrepo := hdl.taskService.GetTaskRepository()
	tasks := taskrepo.GetTaskNames()
	res := &pb.GetTasksResponse{
		Tasks: tasks,
	}
	return res, nil
}
