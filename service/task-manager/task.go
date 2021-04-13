package task_manager

import (
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"github.com/encobrain/go-task-manager/lib/storage/queue"
	"log"
)

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
//
//   req *protocol/mes.CS_TaskStatusSubscribe_rq
//   protocol.ctl protocol/controller.Controller
//   task.status.subscribe.state *taskStatusSubscribeState
func (s *tmService) taskStatusSubscribe(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_TaskStatusSubscribe_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	tsss := ctx.Value("task.status.subscribe.state").(*taskStatusSubscribeState)

	res := &mes.SC_TaskStatusSubscribe_rs{}

	queue := queueGetById(ctx, req.QueueId)

	if queue != nil {
		task := queue.TaskGet(req.UUID)

		if task != nil {
			id := tsss.new()
			res.SubscribeId = &id

			ctx.Child("task.subscribe.process", s.taskStatusSubscribeProcess).
				ValueSet("task.uuid", req.UUID).
				ValueSet("subscribe.id", id).
				ValueSet("queue", queue).Go()
		}
	}

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}

// ctx should contain vars:
//   req *protocol/mes.CS_TaskContent_rq
//   protocol.ctl protocol/controller.Controller
//   task.state *taskState
func (s *tmService) taskContent(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_TaskContent_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	taskState := ctx.Value("task.state").(*taskState)

	task := taskState.getTask(req.StateId)

	res := &mes.SC_TaskContent_rs{}

	if task != nil {
		select {
		case <-task.Canceled():
		default:
			content := task.Content()
			content = append([]byte{}, content...)
			res.Content = &content
		}
	}

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}

// ctx should contain vars:
//   req *protocol/mes.CS_TaskStatusSet_rq
//   protocol.ctl protocol/controller.Controller
//   task.state *taskState
//   storage.queue.manager lib/storage/queue.Manager
func (s *tmService) taskStatusSet(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_TaskStatusSet_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	taskState := ctx.Value("task.state").(*taskState)
	queueManager := ctx.Value("storage.queue.manager").(queue.Manager)

	res := &mes.SC_TaskStatusSet_rs{}

	defer func() {
		err := protCtl.ResponseSend(req, res)

		if err != nil {
			log.Printf("Send response fail. %s\n", err)
		}
	}()

	task := taskState.getTask(req.StateId)

	if task == nil {
		return
	}

	queue := queueManager.Get(req.QueueId)

	if queue == nil {
		return
	}

	select {
	case <-task.Canceled():
		return
	default:
	}

	task = queue.TaskStatusSet(task.UUID(), req.Status, req.Content)

	if task == nil {
		return
	}

	s.task.router.Route(queue, task)

	stateId := taskState.getOrNewId(task)
	res.StateId = &stateId
}

// ctx should contain vars:
//   req *protocol/mes.CS_TaskRemove_rq
//   protocol.ctl protocol/controller.Controller
//   task.state *taskState
//   storage.queue.manager lib/storage/queue.Manager
func (s *tmService) taskRemove(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_TaskRemove_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	taskState := ctx.Value("task.state").(*taskState)
	queueManager := ctx.Value("storage.queue.manager").(queue.Manager)

	res := &mes.SC_TaskRemove_rs{}

	defer func() {
		err := protCtl.ResponseSend(req, res)

		if err != nil {
			log.Printf("Send response fail. %s\n", err)
		}
	}()

	task := taskState.getTask(req.StateId)

	if task == nil {
		return
	}

	queue := queueManager.Get(req.QueueId)

	if queue == nil {
		return
	}

	select {
	case <-task.Canceled():
		return
	default:
	}

	res.Ok = queue.TaskRemove(task.UUID())
}

// ctx should contain vars:
//   req *protocol/mes.CS_TaskReject_rq
//   protocol.ctl protocol/controller.Controller
//   task.state *taskState
//   queue.subscribe.state *queueSubscribeState
func (s *tmService) taskReject(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_TaskReject_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	taskState := ctx.Value("task.state").(*taskState)
	qss := ctx.Value("queue.subscribe.state").(*queueSubscribeState)

	res := &mes.SC_TaskReject_rs{}

	defer func() {
		err := protCtl.ResponseSend(req, res)

		if err != nil {
			log.Printf("Send response fail. %s\n", err)
		}
	}()

	task := taskState.getTask(req.StateId)

	if task != nil {
		qss.rejectSent(task)
	}
}
