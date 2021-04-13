package task_manager

import (
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"github.com/encobrain/go-task-manager/lib/storage"
	"github.com/encobrain/go-task-manager/lib/storage/queue"
	"log"
)

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
func queueGetById(ctx context.Context, id uint64) storage.Queue {
	queueManager := ctx.Value("storage.queue.manager").(queue.Manager)

	return queueManager.Get(id)
}

// ctx should contain vars:
//   req *protocol/mes/CS_ClientGetQueue_rq
//   protocol.ctl protocol/controller.Controller
//   storage.queue.manager lib/storage/queue.Manager
func (s *tmService) queueGet(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_ClientGetQueue_rq)
	queueManager := ctx.Value("storage.queue.manager").(queue.Manager)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)

	res := &mes.SC_ClientGetQueue_rs{}

	queue := queueManager.GetOrCreate(req.Name)

	res.QueueId = queue.ID()

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
//
//   req *protocol/mes/mes.CS_QueueTaskNew_rq
//   task.state *taskState
func (s *tmService) queueTaskNew(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_QueueTaskNew_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	taskState := ctx.Value("task.state").(*taskState)

	res := &mes.SC_QueueTaskNew_rs{}
	queue := queueGetById(ctx, req.QueueId)

	if queue != nil {
		task := queue.TaskNew(req.ParentUUID, req.Status, req.Content)

		s.task.router.Route(queue, task)

		stateId := taskState.getOrNewId(task)

		res.UUID = task.UUID()
		res.StateId = stateId
	}

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
//
//   protocol.ctl protocol/controller.Controller
//   req *protocol/mes/CS_QueueTaskGet_rq
//   task.state *taskState
func (s *tmService) queueTaskGet(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_QueueTaskGet_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	taskState := ctx.Value("task.state").(*taskState)

	res := &mes.SC_QueueTaskGet_rs{}

	queue := queueGetById(ctx, req.QueueId)

	if queue != nil {
		task := queue.TaskGet(req.UUID)

		if task != nil {
			stateId := taskState.getOrNewId(task)

			res.Info = &mes.TaskInfo{
				StateId:    stateId,
				UUID:       task.UUID(),
				ParentUUID: task.ParentUUID(),
				Status:     task.Status(),
			}
		}
	}

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
//   task.state *taskState
//
//   queue.subscribe.state *queueSubscribeState
//   req *protocol/mes/CS_QueueTasksSubscribe_rq
//   protocol.ctl protocol/controller.Controller
func (s *tmService) queueTaskSubscribe(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_QueueTasksSubscribe_rq)
	qss := ctx.Value("queue.subscribe.state").(*queueSubscribeState)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)

	queue := queueGetById(ctx, req.QueueId)

	res := &mes.SC_QueueTasksSubscribe_rs{}

	if queue != nil {
		id := qss.new()
		res.SubscribeId = &id

		receive := s.task.router.Subscribe(queue, req.ParentUUID)

		ctx.Child("queue.subscribe.process", s.queueSubscribeProcess).
			ValueSet("receive", receive).
			ValueSet("subscribe.id", id).
			ValueSet("subscribe.queue", queue).Go()
	}

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
//
//   req *protocol/mes/CS_QueueTasksGet_rq
//   protocol.ctl protocol/controller.Controller
//   task.state *taskState
func (s *tmService) queueTasksGet(ctx context.Context) {
	req := ctx.Value("req").(*mes.CS_QueueTasksGet_rq)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	taskState := ctx.Value("task.state").(*taskState)

	res := &mes.SC_QueueTasksGet_rs{}

	queue := queueGetById(ctx, req.QueueId)

	if queue != nil {
		tasks := queue.TasksGet()

		if req.ParentUUID != "" {
			var byPID []storage.Task

			for _, t := range tasks {
				if t.ParentUUID() == req.ParentUUID {
					byPID = append(byPID, t)
				}
			}

			tasks = byPID
		}

		for _, t := range tasks {
			stateId := taskState.getOrNewId(t)

			res.Tasks = append(res.Tasks, mes.TaskInfo{
				StateId:    stateId,
				UUID:       t.UUID(),
				ParentUUID: t.ParentUUID(),
				Status:     t.Status(),
			})
		}
	}

	err := protCtl.ResponseSend(req, res)

	if err != nil {
		log.Printf("Send response fail. %s\n", err)
	}
}
