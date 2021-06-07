package task_manager

import (
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"github.com/encobrain/go-task-manager/lib/storage"
	"log"
	"sync"
)

func newTaskStatusSubscribeState() *taskStatusSubscribeState {
	return &taskStatusSubscribeState{
		cancels: map[uint64]chan struct{}{},
	}
}

type taskStatusSubscribeState struct {
	mu      sync.Mutex
	nextId  uint64
	cancels map[uint64]chan struct{}
}

func (s *taskStatusSubscribeState) new() (id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id = s.nextId
	s.nextId++

	s.cancels[id] = make(chan struct{})
	return
}

func (s *taskStatusSubscribeState) cancel(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cancel := s.cancels[id]

	if cancel != nil {
		delete(s.cancels, id)
		close(cancel)
	}
}

func (s *taskStatusSubscribeState) getCancel(id uint64) <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.cancels[id]
}

// ctx should contain vars:
//   task.uuid string
//   task.state *taskState
//   queue lib/storage.Queue
//   subscribe.id uint64
//   protocol.ctl protocol/controller.Controller
//   task.status.subscribe.state *taskStatusSubscribeState
func (s *tmService) taskStatusSubscribeProcess(ctx context.Context) {
	taskUUID := ctx.Value("task.uuid").(string)
	taskState := ctx.Value("task.state").(*taskState)
	queue := ctx.Value("queue").(storage.Queue)
	subId := ctx.Value("subscribe.id").(uint64)
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)
	tsss := ctx.Value("task.status.subscribe.state").(*taskStatusSubscribeState)
	cancel := tsss.getCancel(subId)

	defer tsss.cancel(subId)

	for {
		task := queue.TaskGet(taskUUID)

		statusMes := &mes.SC_TaskStatus_ms{
			SubscribeId: subId,
		}

		var wCtx context.Context

		if task != nil {
			var stateId uint64

			stateId, wCtx = taskState.getOrNewId(task)

			statusMes.Info = &mes.TaskInfo{
				StateId:    stateId,
				UUID:       taskUUID,
				ParentUUID: task.ParentUUID(),
				Status:     task.Status(),
			}
		}

		err := protCtl.MessageSend(statusMes)

		if wCtx != nil {
			wCtx.Go()
		}

		if err != nil {
			log.Printf("Task[%s]: Send task status fail. %s\n", taskUUID, err)
			return
		}

		if task == nil {
			log.Printf("Task[%s]: Send task status stopped. Task not exists\n", taskUUID)
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-protCtl.Finished():
			return
		case <-cancel:
			log.Printf("Task[%s]: Send task status canceled\n", taskUUID)
			return
		case <-task.Canceled():
		}
	}
}
