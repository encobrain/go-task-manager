package task_manager

import (
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"log"
)

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
//
//   req *protocol/mes/CS_TaskStatusSubscribe_rq
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

			procCtx := ctx.Child("task.subscribe.process", s.taskStatusSubscribeProcess)
			procCtx.ValueSet("task.uuid", req.UUID)
			procCtx.ValueSet("subscribe.id", id)
			procCtx.ValueSet("queue", queue)
		}
	}

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}

// ctx should contain vars:
//   req *protocol/mes/CS_TaskContent_rq
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
