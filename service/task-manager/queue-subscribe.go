package task_manager

import (
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"github.com/encobrain/go-task-manager/lib/storage"
	"log"
	"sync"
)

func newQueueSubscribeState() *queueSubscribeState {
	return &queueSubscribeState{
		cancels: map[uint64]chan struct{}{},
	}
}

type queueSubscribeState struct {
	mu      sync.Mutex
	nextId  uint64
	cancels map[uint64]chan struct{}
}

func (s *queueSubscribeState) new() (id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id = s.nextId
	s.nextId++

	s.cancels[id] = make(chan struct{})
	return
}

func (s *queueSubscribeState) cancel(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cancel := s.cancels[id]

	if cancel != nil {
		delete(s.cancels, id)
		close(cancel)
	}
}

func (s *queueSubscribeState) getCancel(id uint64) <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.cancels[id]
}

// ctx should contain vars:
//   subscribe.queue lib/storage.Queue
//   task.state *taskState
//
//   queue.subscribe.state *queueSubscribeState
//   protocol.ctl protocol/controller.Controller
//   receive <-chan storage.Task
//   subscribe.id uint64
func (s *tmService) queueSubscribeProcess(ctx context.Context) {
	receive := ctx.Value("receive").(<-chan storage.Task)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	qss := ctx.Value("queue.subscribe.state").(*queueSubscribeState)
	subId := ctx.Value("subscribe.id").(uint64)

	defer qss.cancel(subId)

	for {
		select {
		case <-ctx.Done():
			return
		case <-protCtl.Finished():
			return
		case task := <-receive:
			ctx := ctx.Child("send", s.queueSubscribeSend)
			ctx.ValueSet("task", task)
		}
	}
}

// ctx should contain vars:
//   task.state *taskState
//
//   task lib/storage.Task
//   protocol.ctl protocol/controller.Controller
//   subscribe.id uint64
//   subscribe.queue lib/storage.Queue
//   queue.subscribe.state *queueSubscribeState
func (s *tmService) queueSubscribeSend(ctx context.Context) {
	task := ctx.Value("task").(storage.Task)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	qss := ctx.Value("queue.subscribe.state").(*queueSubscribeState)
	subId := ctx.Value("subscribe.id").(uint64)
	cancel := qss.getCancel(subId)
	queue := ctx.Value("subscribe.queue").(storage.Queue)

	defer s.task.router.Route(queue, task)

	if cancel == nil {
		return
	}

	stateId := s.taskStateGetOrNewId(ctx, task)

	defer s.taskStateRemove(ctx, task)

	err := protCtl.MessageSend(&mes.SC_QueueSubscribeTask_ms{
		SubscribeId: subId,
		Info: mes.TaskInfo{
			StateId:    stateId,
			UUID:       task.UUID(),
			ParentUUID: task.ParentUUID(),
			Status:     task.Status(),
		},
	})

	if err != nil {
		log.Printf("Send subscribe task fail. %s\n", err)

		return
	}

	select {
	case <-ctx.Done():
	case <-task.Canceled():
	case <-cancel:
	case <-protCtl.Finished():
	}
}
