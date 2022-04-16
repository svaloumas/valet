package main

import (
	"github.com/svaloumas/valet"
	"github.com/svaloumas/valet/task"
)

func main() {
	v := valet.New("config.yaml")
	v.RegisterTask("dummytask", task.DummyTask)
	v.Run()
}
