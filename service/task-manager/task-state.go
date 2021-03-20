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

func newTaskState() *taskState {
	return &taskState{
		id:      map[<-chan struct{}]uint64{},
		watcher: map[<-chan struct{}]context.Context{},
		task:    map[uint64]storage.Task{},
	}
}

type taskState struct {
	mu      sync.Mutex
	nextId  uint64
	id      map[<-chan struct{}]uint64
	watcher map[<-chan struct{}]context.Context
	task    map[uint64]storage.Task
}

func (ts *taskState) getTask(stateId uint64) storage.Task {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.task[stateId]
}

// ctx should contain vars:
//   protocol.ctl protocol/controller.Controller
//
//   task.state *taskState
func (s *tmService) taskStateGetOrNewId(ctx context.Context, task storage.Task) (id uint64) {
	taskState := ctx.Value("task.state").(*taskState)
	taskState.mu.Lock()
	defer taskState.mu.Unlock()

	chId := task.Canceled()
	id, ok := taskState.id[chId]

	if ok {
		return id
	}

	id = taskState.nextId
	taskState.nextId++

	taskState.id[chId] = id
	taskState.task[id] = task

	wCtx := ctx.Child("task.state.watcher", s.taskStateWatcher)
	wCtx.ValueSet("task", task)

	return
}

// ctx should contain vars:
//   task.state *taskState
//   protocol.ctl protocol/controller.Controller
//   task lib/storage.Task
func (s *tmService) taskStateWatcher(ctx context.Context) {
	task := ctx.Value("task").(storage.Task)
	chId := task.Canceled()

	taskState := ctx.Value("task.state").(*taskState)
	taskState.mu.Lock()

	id, ok := taskState.id[chId]

	if !ok {
		taskState.mu.Unlock()
		return
	}

	protCtl := ctx.Value("protocol.ctl").(controller.Controller)

	taskState.watcher[chId] = ctx

	taskState.mu.Unlock()

	defer s.taskStateRemove(ctx, task)

	select {
	case <-ctx.Done():
		return
	case <-protCtl.Finished():
		return
	case <-task.Canceled():
		err := protCtl.MessageSend(&mes.SC_TaskCancel_ms{
			StateId: id,
			Reason:  "state changed",
		})

		if err != nil {
			log.Printf("Send task cancel fail. %s\n", err)
		}
	}
}

// ctx should contain vars:
//   task.state *taskState
func (s *tmService) taskStateRemove(ctx context.Context, task storage.Task) {
	taskState := ctx.Value("task.state").(*taskState)
	chId := task.Canceled()
	taskState.mu.Lock()
	defer taskState.mu.Unlock()

	id := taskState.id[chId]

	delete(taskState.id, chId)
	delete(taskState.task, id)

	ictx, ok := taskState.watcher[chId]

	if ok {
		ictx.(context.Context).Cancel(fmt.Errorf("removed"))
	}

	delete(taskState.watcher, chId)
}
