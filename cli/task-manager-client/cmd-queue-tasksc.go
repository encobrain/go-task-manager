package main

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/lib/client"
	"log"
)

type CmdQueueTasksCount struct {
}

func (c CmdQueueTasksCount) Execute(args []string) (err error) {
	conf.Init()

	if len(args) < 3 {
		return fmt.Errorf("use: queue_tasksc <queue> <parentUUID> <status>")
	}

	queueName := args[0]
	parentUUID := args[1]
	status := args[2]

	cl := client.NewClient(context.Main, &conf.Client)
	cl.Start()

	log.Printf("Getting queue `%s`...", queueName)

	queue := <-cl.GetQueue(queueName)

	if queue == nil {
		return fmt.Errorf("get queue fail. client stopped")
	}

	log.Printf("Getting tasks parentUUID=`%s` status=`%s`...", parentUUID, status)

	tasks := <-queue.TasksGet(parentUUID)

	if tasks == nil {
		return fmt.Errorf("get tasks fail. client stopped")
	}

	if status != "" {
		var f []client.Task

		for _, t := range tasks {
			if t.Status() == status {
				f = append(f, t)
			}
		}
	}

	log.Printf("Got %d tasks", len(tasks))

	return nil
}

func (c CmdQueueTasksCount) Name() string {
	return "queue_tasksc"
}

func (c CmdQueueTasksCount) Description() string {
	return "Count queue's tasks"
}

func (c CmdQueueTasksCount) LongDescription() string {
	return "Count queue's tasks"
}
