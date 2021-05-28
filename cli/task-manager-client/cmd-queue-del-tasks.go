package main

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/lib/client"
	"log"
	"sync"
)

type CmdQueueDelTasks struct {
}

func (c CmdQueueDelTasks) Execute(args []string) (err error) {
	conf.Init()

	if len(args) < 3 {
		return fmt.Errorf("use: queue_del_tasks <queue> <parentUUID> <status>")
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

	wg := sync.WaitGroup{}
	wg.Add(len(tasks))

	for _, t := range tasks {
		t := t

		go func() {
			<-t.Remove()
			wg.Done()
		}()
	}

	wg.Wait()

	log.Printf("Done")

	return nil
}

func (c CmdQueueDelTasks) Name() string {
	return "queue_del_tasks"
}

func (c CmdQueueDelTasks) Description() string {
	return "Deletes queue's tasks"
}

func (c CmdQueueDelTasks) LongDescription() string {
	return "Deletes queue's tasks"
}
