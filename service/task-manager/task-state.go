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

// ctx should contain vars:
//   protocol.ctl protocol/controller.Controller
func newTaskState(ctx context.Context) *taskState {
	return &taskState{
		ctx:      ctx,
		ids:      map[<-chan struct{}]uint64{},
		watchers: map[<-chan struct{}]context.Context{},
		tasks:    map[uint64]storage.Task{},
	}
}

type taskState struct {
	ctx     context.Context
	protCtl controller.Controller

	mu       sync.Mutex
	nextId   uint64
	ids      map[<-chan struct{}]uint64
	watchers map[<-chan struct{}]context.Context
	tasks    map[uint64]storage.Task
}

func (ts *taskState) getOrNewId(task storage.Task) (id uint64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	chId := task.Canceled()
	id, ok := ts.ids[chId]

	if ok {
		return id
	}

	id = ts.nextId
	ts.nextId++

	ts.ids[chId] = id
	ts.tasks[id] = task

	wCtx := ts.ctx.Child("task.state.watcher", ts.watcher).
		ValueSet("task", task).Go()

	ts.watchers[chId] = wCtx

	return
}

// ctx should contain vars:
//   protocol.ctl protocol/controller.Controller
//   task lib/storage.Task
func (ts *taskState) watcher(ctx context.Context) {
	task := ctx.Value("task").(storage.Task)
	chId := task.Canceled()
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)

	ts.mu.Lock()

	id, ok := ts.ids[chId]

	ts.mu.Unlock()

	if !ok {
		return
	}

	defer ts.remove(task)

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

func (ts *taskState) remove(task storage.Task) {
	chId := task.Canceled()
	ts.mu.Lock()
	defer ts.mu.Unlock()

	id := ts.ids[chId]

	delete(ts.ids, chId)
	delete(ts.tasks, id)

	wCtx := ts.watchers[chId]

	if wCtx != nil {
		wCtx.Cancel(fmt.Errorf("removed"))
	}

	delete(ts.watchers, chId)
}

func (ts *taskState) getTask(stateId uint64) storage.Task {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.tasks[stateId]
}
