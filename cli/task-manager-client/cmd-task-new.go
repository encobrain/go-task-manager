package main

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/lib/client"
	"log"
)

type CmdTaskAdd struct {
}

func (c CmdTaskAdd) Execute(args []string) (err error) {
	conf.Init()

	if len(args) < 4 {
		return fmt.Errorf("use: task_add <queue> <parentUUID> <status> <content>")
	}

	queueName := args[0]
	parentUUID := args[1]
	status := args[2]
	content := args[3]

	client := client.NewClient(context.Main, &conf.Client)
	client.Start()

	log.Printf("Getting queue `%s`...", queueName)

	queue := <-client.GetQueue(queueName)

	if queue == nil {
		return fmt.Errorf("get queue fail. client stopped")
	}

	log.Printf("Creating task parentUUID=`%s` status=`%s` content=`%s`...", parentUUID, status, content)

	task := <-queue.TaskNew(parentUUID, status, []byte(content))

	if task == nil {
		return fmt.Errorf("create task fail. client stopped")
	}

	log.Printf("Task created")

	return nil
}

func (c CmdTaskAdd) Name() string {
	return "task_add"
}

func (c CmdTaskAdd) Description() string {
	return "Adds task to queue"
}

func (c CmdTaskAdd) LongDescription() string {
	return "Adds task to queue"
}
