package main

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/lib/client"
	"log"
)

type CmdTaskDel struct {
}

func (c CmdTaskDel) Execute(args []string) (err error) {
	conf.Init()

	if len(args) < 2 {
		return fmt.Errorf("use: task_del <queue> <uuid>")
	}

	queueName := args[0]
	uuid := args[1]

	client := client.NewClient(context.Main, &conf.Client)
	client.Start()

	log.Printf("Getting queue `%s`...", queueName)

	queue := <-client.GetQueue(queueName)

	if queue == nil {
		return fmt.Errorf("get queue fail. client stopped")
	}

	log.Printf("Getting task UUID=`%s`...", uuid)

	task := <-queue.TaskGet(uuid)

	if task == nil {
		return fmt.Errorf("get task fail. task not exists")
	}

	log.Printf("Removing task...")

	<-task.Remove()

	log.Printf("Task removed")

	return nil
}

func (c CmdTaskDel) Name() string {
	return "task_del"
}

func (c CmdTaskDel) Description() string {
	return "Deletes task from queue"
}

func (c CmdTaskDel) LongDescription() string {
	return "Deletes task from queue"
}
