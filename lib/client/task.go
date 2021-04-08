package client

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"log"
	"sync"
)

// Task in any moment may be canceled. See Canceled()
type Task interface {
	UUID() string
	ParentUUID() string
	// Status gets status of task.
	Status() string
	// StatusSubscribe subscribes on actual task status from queue.
	// If task state changed - previous task will be canceled.
	// No need listen to task.Canceled() and status simultaniously.
	// Empty status - task not exists.
	// Closed chan - client stopped
	StatusSubscribe() (status <-chan Task)
	// Content gets tasks content from task-manager.
	// nil - task canceled.
	// Chan closed with result.
	Content() (content <-chan []byte)
	// StatusSet sets status of task and new content.
	// If <-done == false - task canceled
	StatusSet(status string, content []byte) (done <-chan bool)
	// Remove removes task from queue.
	// Remove should execute on side that created this task.
	// If <-done == false - task canceled
	Remove() (done <-chan bool)
	// Canceled returns canceled channel.
	// If it closed - task canceled. See Err().
	Canceled() (canceled <-chan struct{})
	// Err return reason of task cancel.
	Err() (err error)
	// Reject rejects task for to do by another worker.
	// Method should use if task received by queue.TasksSubscribe().
	Reject() (done <-chan struct{})
}

func taskNew() *task {
	return &task{
		canceled: make(chan struct{}),
	}
}

type task struct {
	ctx context.Context

	protocol struct {
		ctl chan controller.Controller
	}

	stateId    uint64
	uuid       string
	parentUUID string
	status     string

	mu       sync.Mutex
	canceled chan struct{}
	err      error
}

func (t *task) cancel(reason error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.err != nil {
		return
	}

	t.err = reason

	close(t.canceled)
}

func (t *task) UUID() string {
	return t.uuid
}

func (t *task) ParentUUID() string {
	return t.parentUUID
}

func (t *task) Status() string {
	return t.status
}

func (t *task) StatusSubscribe() (status <-chan Task) {

}

func (t *task) Content() (content <-chan []byte) {
	ch := make(chan []byte, 1)

	t.ctx.Child("task.content", func(ctx context.Context) {
		defer close(ch)

		log.Printf("TMClient: Task[%s]: getting content...", t.uuid)

		for {
			var protCtl controller.Controller

			select {
			case <-t.canceled:
				return
			case protCtl = <-t.protocol.ctl:
			}

			if protCtl == nil {
				log.Printf("TMClient: Task[%s]: client stopped\n", t.uuid)
				return
			}

			res, err := protCtl.RequestSend(&mes.CS_TaskContent_rq{
				StateId: t.stateId,
			})

			if err != nil {
				log.Printf("TMClient: Task[%s]: send request fail. %s\n", t.uuid, err)
				continue
			}

			select {
			case <-t.ctx.Done():
				return
			case <-t.canceled:
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_TaskContent_rs)

				if rs.Content == nil {
					return
				}

				log.Printf("TMClient: Task[%s]: got content\n", t.uuid)

				ch <- *rs.Content
				return
			}

		}
	})

	return ch
}

func (t *task) StatusSet(status string, content []byte) (done <-chan bool) {

}

func (t *task) Remove() (done <-chan bool) {

}

func (t *task) Canceled() (canceled <-chan struct{}) {
	return t.canceled
}

func (t *task) Err() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.err
}

func (t *task) Reject() (done <-chan struct{}) {

}
