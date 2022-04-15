package main

import (
	"github.com/svaloumas/valet/task"
	"github.com/svaloumas/valet/valet"
)

func main() {
	taskrepo := valet.NewTaskRepository()
	taskrepo.Register("dummytask", task.DummyTask)
	valet.Run("config.yaml", taskrepo)
}
