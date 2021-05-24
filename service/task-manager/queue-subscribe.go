package task_manager

import (
	"fmt"
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
		sent:    map[storage.Task]context.Context{},
	}
}

type queueSubscribeState struct {
	mu      sync.Mutex
	nextId  uint64
	cancels map[uint64]chan struct{}
	sent    map[storage.Task]context.Context
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

func (s *queueSubscribeState) addSent(task storage.Task, ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sent[task] = ctx
}

func (s *queueSubscribeState) rejectSent(task storage.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := s.sent[task]

	if ctx != nil {
		delete(s.sent, task)
		ctx.Cancel(fmt.Errorf("rejected"))
	}
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
	cancel := qss.getCancel(subId)

	defer qss.cancel(subId)

	for {
		select {
		case <-ctx.Done():
			return
		case <-protCtl.Finished():
			return
		case <-cancel:
			return
		case task := <-receive:
			ctx.Child("send", s.queueSubscribeSend).
				ValueSet("task", task).Go()
		}
	}
}

// ctx should contain vars:
//   task lib/storage.Task
//   task.state *taskState
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
	taskState := ctx.Value("task.state").(*taskState)

	defer s.task.router.Route(queue, task)

	if cancel == nil {
		return
	}

	stateId := taskState.getOrNewId(task)

	qss.addSent(task, ctx)

	defer qss.rejectSent(task)

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
