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

	if len(args) < 2 {
		return fmt.Errorf("use: queue_tasksc <queue> <parentUUID>")
	}

	queueName := args[0]
	parentUUID := args[1]

	cl := client.NewClient(context.Main, &conf.Client)
	cl.Start()

	log.Printf("Getting queue `%s`...", queueName)

	queue := <-cl.GetQueue(queueName)

	if queue == nil {
		return fmt.Errorf("get queue fail. client stopped")
	}

	log.Printf("Getting tasks parentUUID=`%s`...", parentUUID)

	tasks := <-queue.TasksGet(parentUUID)

	if tasks == nil {
		return fmt.Errorf("get tasks fail. client stopped")
	}

	counts := map[string]int{}

	log.Printf("Got %d tasks", len(tasks))

	for _, t := range tasks {
		counts[t.Status()]++
	}

	for status, count := range counts {
		log.Printf("%s: %d", status, count)
	}

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
