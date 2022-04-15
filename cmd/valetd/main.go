package main

import (
	"valet/task"
	"valet/valet"
)

func main() {
	taskrepo := valet.NewTaskRepository()
	taskrepo.Register("dummytask", task.DummyTask)
	valet.Run("config.yaml", taskrepo)
}
