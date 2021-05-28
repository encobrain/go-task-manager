package main

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/lib/client"
	"log"
	"sync"
)

type CmdQueueTasks struct {
}

func (c CmdQueueTasks) Execute(args []string) (err error) {
	conf.Init()

	if len(args) < 3 {
		return fmt.Errorf("use: queue_tasks <queue> <parentUUID> <status>")
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

		tasks = f
	}

	log.Printf("Got %d tasks", len(tasks))

	cont := map[client.Task][]byte{}

	wg := sync.WaitGroup{}
	wg.Add(len(tasks))

	for _, t := range tasks {
		t := t

		go func() {
			bytes := <-t.Content()

			cont[t] = bytes
			wg.Done()
		}()
	}

	wg.Wait()

	for _, t := range tasks {
		log.Printf("UUID: %s PUUID: %s STATUS: %s CONTENT: %s", t.UUID(), t.ParentUUID(), t.Status(), string(cont[t]))
	}

	return nil
}

func (c CmdQueueTasks) Name() string {
	return "queue_tasks"
}

func (c CmdQueueTasks) Description() string {
	return "List queue's tasks"
}

func (c CmdQueueTasks) LongDescription() string {
	return "List queue's tasks"
}
