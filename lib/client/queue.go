package client

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"log"
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
	// TasksSubscribe subscribes to get tasks from queue for process.
	// If paretnUUID not empty - subscribe onlyon tasks with that parent uuid
	// If nil - client stopped.
	TasksSubscribe(parentUUID string) (tasks <-chan Task)
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

		log.Printf("TMClient: Queue[%d]: creating new task: parentUUID=%s status=%s ...\n", q.id, parentUUID, status)

		for {
			protCtl := <-q.protocol.ctl

			if protCtl == nil {
				log.Printf("TMClient: Queue[%d]: client stopped\n", q.id)
				return
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTaskNew_rq{
				QueueId:    q.id,
				ParentUUID: parentUUID,
				Status:     status,
				Content:    content,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%d]: send request fail. %s\n", q.id, err)
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

				log.Printf("TMClient: Queue[%d]: created task UUID=%s\n", q.id, rs.UUID)

				t := q.tasks.new(q.id, rs.StateId, rs.UUID, parentUUID, status)

				ch <- t
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

		log.Printf("TMClient: Queue[%d]: getting task UUID=%s ...\n", q.id, uuid)

		for {
			protCtl := <-q.protocol.ctl

			if protCtl == nil {
				log.Printf("TMClient: Queue[%d]: client stopped\n", q.id)
				return
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTaskGet_rq{
				QueueId: q.id,
				UUID:    uuid,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%d]: send request fail. %s\n", q.id, err)
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

				log.Printf("TMClient: Queue[%d]: got task UUID=%s\n", q.id, rs.Info.UUID)

				t := q.tasks.new(q.id, rs.Info.StateId, rs.Info.UUID, rs.Info.ParentUUID, rs.Info.Status)

				ch <- t
				return
			}
		}
	}).Go()

	return ch
}

func (q *queue) TasksSubscribe(parentUUID string) (tasks <-chan Task) {
	ch := make(chan Task)

	q.ctx.Child("queue.tasks.subscribe", func(ctx context.Context) {
		log.Printf("TMClient: Queue[%d]: tasks subscribing parentUUID=%s ...\n", q.id, parentUUID)

		for {
			protCtl := <-q.protocol.ctl

			if protCtl == nil {
				log.Printf("TMClient: Queue[%d]: client stopped\n", q.id)
				close(ch)
				return
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTasksSubscribe_rq{
				QueueId:    q.id,
				ParentUUID: parentUUID,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%d]: send request fail. %s\n", q.id, err)
				continue
			}

			select {
			case <-q.ctx.Done():
				close(ch)
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_QueueTasksSubscribe_rs)

				if rs.SubscribeId == nil {
					panic(fmt.Errorf("tasks subscribe fail. queueId[%d] invalid", q.id))
				}

				log.Printf("TMClient: Queue[%d]: subscribe done. subscribeId=%d\n", q.id, *rs.SubscribeId)

				q.tasks.subscribe(*rs.SubscribeId, q.id, ch)
			}

			select {
			case <-ctx.Done():
				return
			case <-protCtl.Finished():
			}

			log.Printf("TMClient: Queue[%d]: tasks resubscribing parentUUID=%s ...\n", q.id, parentUUID)
		}
	}).Go()

	return ch
}

func (q *queue) TasksGet(parentUUID string) (tasks <-chan []Task) {
	ch := make(chan []Task)

	q.ctx.Child("queue.tasks.get", func(ctx context.Context) {
		defer close(ch)

		log.Printf("TMClient: Queue[%d]: getting tasks ParentUUID=%s ...\n", q.id, parentUUID)

		for {
			protCtl := <-q.protocol.ctl

			if protCtl == nil {
				log.Printf("TMClient: Queue[%d]: client stopped\n", q.id)
				return
			}

			res, err := protCtl.RequestSend(&mes.CS_QueueTasksGet_rq{
				QueueId:    q.id,
				ParentUUID: parentUUID,
			})

			if err != nil {
				log.Printf("TMClient: Queue[%d]: send request fail. %s\n", q.id, err)
				continue
			}

			select {
			case <-q.ctx.Done():
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_QueueTasksGet_rs)

				log.Printf("TMClient: Queue[%d]: got tasks parentUUID=%s count=%d\n", q.id, parentUUID, len(rs.Tasks))

				var tasks []Task

				for _, i := range rs.Tasks {
					tasks = append(tasks, q.tasks.new(q.id, i.StateId, i.UUID, i.ParentUUID, i.Status))
				}

				ch <- tasks
				return
			}
		}
	}).Go()

	return ch
}
