package client

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"log"
	"runtime"
	"strings"
)

type Queue interface {
	// Name returns queue name
	Name() string
	// TaskNew creates new task with parrentUUID. parentUUID may be empty.
	// Chan closed with result.
	// If nil - client stopped.
	TaskNew(parentUUID string, status string, content []byte) (task <-chan Task)
	// TaskGet gets task with uuid.
	// Chan closed with result.
	// If nil - task not exists.
	TaskGet(uuid string) (task <-chan Task)
	// TasksSubscribe subscribes to get tasks from queue for process by parentUUID & status.
	// Empty values means Any.
	// If nil - client stopped.
	// Close tasks channel - cancel subscribe
	TasksSubscribe(parentUUID string, status string) (tasks chan Task)
	// TasksGet gets all tasks with/or without parentUUID in queue.
	// If nil - client stopped.
	// Chan closed with result.
	TasksGet(parentUUID string) (tasks <-chan []Task)
}

func newQueue() *queue {
	return &queue{}
}

type queue struct {
	id   uint64
	name string
	ctx  context.Context

	tasks struct {
		new       func(queueId uint64, stateId uint64, uuid string, parentUUID string, status string) *task
		subscribe func(subscribeId uint64, queueId uint64, ch chan Task)
	}

	protocol struct {
		ctl chan controller.Controller
	}
}

func (q *queue) Name() string {
	return q.name
}

func (q *queue) TaskNew(parentUUID string, status string, content []byte) (task <-chan Task) {
	ch := make(chan Task, 1)

	q.ctx.Child("queue.task.new", func(ctx context.Context) {
		defer close(ch)

		log.Printf("TMClient: Queue[%s]: creating new task: parentUUID=%s status=%s ...\n", q.name, parentUUID, status)

		for {
			var protCtl controller.Controller

			select {
			case <-ctx.Done():
				return
			case protCtl = <-q.protocol.ctl:
				if protCtl == nil {
					log.Printf("TMClient: Queue[%s]: client stopped\n", q.name)
					return
				}
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTaskNew_rq{
				QueueId:    q.id,
				ParentUUID: parentUUID,
				Status:     status,
				Content:    content,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%s]: send request fail. %s\n", q.name, err)
				continue
			}

			select {
			case <-q.ctx.Done():
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_QueueTaskNew_rs)

				if rs.UUID == "" {
					panic(fmt.Errorf("create task fail. queueId[%d] invalid", q.id))
				}

				log.Printf("TMClient: Queue[%s]: created task UUID=%s\n", q.name, rs.UUID)

				t := q.tasks.new(q.id, rs.StateId, rs.UUID, parentUUID, status)

				select {
				case <-ctx.Done():
				case ch <- t:
				}

				return
			}

		}
	}).Go()

	return ch
}

func (q *queue) TaskGet(uuid string) (task <-chan Task) {
	ch := make(chan Task)

	q.ctx.Child("queue.task.get", func(ctx context.Context) {
		defer close(ch)

		log.Printf("TMClient: Queue[%s]: getting task UUID=%s ...\n", q.name, uuid)

		for {
			var protCtl controller.Controller

			select {
			case <-ctx.Done():
				return
			case protCtl = <-q.protocol.ctl:
				if protCtl == nil {
					log.Printf("TMClient: Queue[%s]: client stopped\n", q.name)
					return
				}
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTaskGet_rq{
				QueueId: q.id,
				UUID:    uuid,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%s]: send request fail. %s\n", q.name, err)
				continue
			}

			select {
			case <-q.ctx.Done():
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_QueueTaskGet_rs)

				if rs.Info == nil {
					return
				}

				log.Printf("TMClient: Queue[%s]: got task UUID=%s\n", q.name, rs.Info.UUID)

				t := q.tasks.new(q.id, rs.Info.StateId, rs.Info.UUID, rs.Info.ParentUUID, rs.Info.Status)

				select {
				case <-ctx.Done():
				case ch <- t:
				}

				return
			}
		}
	}).Go()

	return ch
}

func (q *queue) TasksSubscribe(parentUUID string, status string) (tasks chan Task) {
	ch := make(chan Task)

	q.ctx.Child("queue.tasks.subscribe", func(ctx context.Context) {
		var e interface{}

		defer func() {
			recover()

			if e == nil {
				return
			}

			if re, ok := e.(runtime.Error); ok {
				if strings.Contains(re.Error(), "send on closed channel") {
					return
				}
			}

			panic(e)
		}()

		defer close(ch)

		defer func() {
			e = recover()
		}()

		var subscribeId *uint64

		defer func() {
			if subscribeId == nil {
				return
			}

			log.Printf("TMClient: Queue[%s]: unsubscribing %d...\n", q.name, *subscribeId)

			for {
				var protCtl controller.Controller

				select {
				case <-ctx.Done():
					return
				case protCtl = <-q.protocol.ctl:
					if protCtl == nil {
						log.Printf("TMClient: Queue[%s]: client stopped\n", q.name)
						return
					}
				}

				res, err := protCtl.RequestSend(&mes.CS_QueueTasksUnsubscribe_rq{
					SubscribeId: *subscribeId,
				})

				if err != nil {
					log.Printf("TMClient: Queue[%s]: send request fail. %s\n", q.name, err)
					continue
				}

				select {
				case <-ctx.Done():
				case resm := <-res:
					if resm != nil {
						log.Printf("TMClient: Queue[%s]: ubsubscribe %d done\n", q.name, *subscribeId)
					}
				}

				return
			}
		}()

		log.Printf("TMClient: Queue[%s]: tasks subscribing parentUUID=%s status=%s ...\n", q.name, parentUUID, status)

		for {
			var protCtl controller.Controller

			select {
			case <-ctx.Done():
				return
			case protCtl = <-q.protocol.ctl:
				if protCtl == nil {
					log.Printf("TMClient: Queue[%s]: client stopped\n", q.name)
					return
				}
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTasksSubscribe_rq{
				QueueId:    q.id,
				ParentUUID: parentUUID,
				Status:     status,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%s]: send request fail. %s\n", q.name, err)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_QueueTasksSubscribe_rs)

				subscribeId = rs.SubscribeId

				if subscribeId == nil {
					panic(fmt.Errorf("tasks subscribe fail. queueId[%d] invalid", q.id))
				}

				log.Printf("TMClient: Queue[%s]: subscribe done. subscribeId=%d\n", q.name, *rs.SubscribeId)

				q.tasks.subscribe(*rs.SubscribeId, q.id, ch)
			}

		send:
			for {
				select {
				case <-ctx.Done():
					subscribeId = nil
					return
				case v := <-ch:
					ch <- v
				case <-protCtl.Finished():
					subscribeId = nil
					break send
				}
			}

			log.Printf("TMClient: Queue[%s]: tasks resubscribing parentUUID=%s status=%s ...\n", q.name, parentUUID, status)
		}
	}).Go()

	return ch
}

func (q *queue) TasksGet(parentUUID string) (tasks <-chan []Task) {
	ch := make(chan []Task)

	q.ctx.Child("queue.tasks.get", func(ctx context.Context) {
		defer close(ch)

		log.Printf("TMClient: Queue[%s]: getting tasks ParentUUID=%s ...\n", q.name, parentUUID)

		for {
			var protCtl controller.Controller

			select {
			case <-ctx.Done():
				return
			case protCtl = <-q.protocol.ctl:
				if protCtl == nil {
					log.Printf("TMClient: Queue[%s]: client stopped\n", q.name)
					return
				}
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTasksGet_rq{
				QueueId:    q.id,
				ParentUUID: parentUUID,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%s]: send request fail. %s\n", q.name, err)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_QueueTasksGet_rs)

				log.Printf("TMClient: Queue[%s]: got tasks parentUUID=%s count=%d\n", q.name, parentUUID, len(rs.Tasks))

				tasks := make([]Task, 0, len(rs.Tasks))

				for _, i := range rs.Tasks {
					tasks = append(tasks, q.tasks.new(q.id, i.StateId, i.UUID, i.ParentUUID, i.Status))
				}

				select {
				case <-ctx.Done():
				case ch <- tasks:
				}

				return
			}
		}
	}).Go()

	return ch
}
