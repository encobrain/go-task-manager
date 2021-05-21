package client

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"github.com/encobrain/go-task-manager/model/config"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"runtime/debug"
	"sync"
	"time"
)

type Client interface {
	// GetQueue get queue. If nil - client stopped.
	// Chan closed with result.
	GetQueue(name string) (queue <-chan Queue)
}

type ClientWithControl interface {
	Client

	Start()
	Stop() (done <-chan struct{})
}

func NewClient(ctx context.Context, config *config.Client) ClientWithControl {
	c := &client{conf: config}
	c.ctx.glob = ctx
	c.protocol.ctl = make(chan controller.Controller)
	c.queue.list = map[string]*queue{}
	c.queue.tasksSubscribe = map[uint64]*subInfo{}
	c.task.list = map[uint64]interface{}{}
	c.task.statusSubscribe = map[uint64]interface{}{}
	close(c.protocol.ctl)

	return c
}

type subInfo struct {
	queueId uint64
	ch      chan Task
}

type client struct {
	mu sync.Mutex

	conf *config.Client

	ctx struct {
		glob   context.Context
		worker context.Context
	}

	queue struct {
		mu             sync.Mutex
		list           map[string]*queue
		tasksSubscribe map[uint64]*subInfo
	}

	task struct {
		mu              sync.Mutex
		list            map[uint64]interface{} // map[stateId]*task|*mes.SC_TaskCancel_ms
		statusSubscribe map[uint64]interface{} // *subInfo|chan struct{}
	}

	protocol struct {
		ctl chan controller.Controller // if closed - client stopped
	}
}

func (c *client) GetQueue(name string) (queue <-chan Queue) {
	ch := make(chan Queue, 1)

	c.ctx.glob.Child("queue.get", func(ctx context.Context) {
		defer close(ch)

		c.queue.mu.Lock()
		queue := c.queue.list[name]
		c.queue.mu.Unlock()

		if queue != nil {
			ch <- queue
			return
		}

		for {
			var protCtl controller.Controller

			select {
			case <-ctx.Done():
				return
			case protCtl = <-c.protocol.ctl:
				if protCtl == nil {
					return
				}
			}

			res, err := protCtl.RequestSend(&mes.CS_ClientGetQueue_rq{
				Name: name,
			})

			if err != nil {
				log.Printf("TMClient: GetQueue(): send request fail. %s\n", err)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case resm := <-res:
				if resm == nil {
					continue
				}

				rs := resm.(*mes.SC_ClientGetQueue_rs)

				q := c.queueNew(name, rs.QueueId)

				ch <- q
				return
			}
		}
	}).Go()

	return ch
}

func (c *client) queueNew(name string, id uint64) *queue {
	c.queue.mu.Lock()
	defer c.queue.mu.Unlock()

	q := c.queue.list[name]

	if q != nil {
		return q
	}

	q = newQueue()
	q.id = id
	q.name = name
	q.tasks.new = c.taskNew
	q.tasks.subscribe = c.queueTasksSubscribe
	q.protocol.ctl = c.protocol.ctl
	q.ctx = c.ctx.worker

	c.queue.list[name] = q

	return q
}

func (c *client) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctx.worker != nil {
		return
	}

	c.protocol.ctl = make(chan controller.Controller)
	c.ctx.worker = c.ctx.glob.Child("worker", c.workerStart).Go()
}

func (c *client) Stop() (done <-chan struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctx.worker == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}

	wctx := c.ctx.worker

	wctx.Cancel(fmt.Errorf("client stopped"))

	done = wctx.Finished(true)

	go func() {
		<-done
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.ctx.worker == wctx {
			c.ctx.worker = nil
		}
	}()

	return
}

func (c *client) workerStart(ctx context.Context) {
	ctx.PanicHandlerSet(func(ctx context.Context, panicVal interface{}) {
		log.Printf("TMClient: worker panic. %s\n", panicVal)
		debug.PrintStack()
		ctx.Cancel(fmt.Errorf("panic"))
	})

	ctx.Child("conn.worker", c.connWorker).Go()

	defer close(c.protocol.ctl)

	<-ctx.Done()

	log.Printf("TMClient: worker stopped. %s\n", ctx.Err())
}

func (c *client) connWorker(ctx context.Context) {
	addr := fmt.Sprintf("%s:%d", c.conf.Connect.Host, c.conf.Connect.Port)
	u := url.URL{Scheme: c.conf.Connect.Scheme, Host: addr, Path: c.conf.Connect.Path}

	var conn *websocket.Conn
	var err error

	defer func() {
		if conn != nil {
			err = conn.Close()

			if err != nil {
				log.Printf("TMClient: close connection fail. %s\n", err)
			} else {
				log.Printf("TMClient: connection closed\n")
			}

			c.connDisconnected()
		}
	}()

	for {
		log.Printf("TMClient: connecting to %s...\n", u.String())

		conn, _, err = websocket.DefaultDialer.DialContext(ctx, u.String(), nil)

		if err != nil {
			log.Printf("TMClient: connect fail. %s\n", err)

			select {
			case <-ctx.Done():
				return
			default:
			}

			log.Printf("TMClient: retry connect after 1sec...\n")

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}

		log.Printf("TMClient: connected to %s\n", u.String())

		protCtl := controller.New(mes.Messages, conn)

		ctx.Child("conn.read", c.connRead).
			ValueSet("protocol.ctl", protCtl).Go()

	wait:
		for {
			select {
			case <-ctx.Done():
				return
			case <-protCtl.Finished():
				c.connDisconnected()
				break wait
			case c.protocol.ctl <- protCtl:
			}
		}
	}
}

func (c *client) connDisconnected() {
	c.task.mu.Lock()
	defer c.task.mu.Unlock()

	for _, itask := range c.task.list {
		t, ok := itask.(*task)

		if ok {
			t.cancel(fmt.Errorf("client disconnected"))
		}
	}

	c.task.list = map[uint64]interface{}{}

	c.task.statusSubscribe = map[uint64]interface{}{}

	c.queue.mu.Lock()
	defer c.queue.mu.Unlock()

	c.queue.tasksSubscribe = map[uint64]*subInfo{}
}

// ctx should contain vars:
//   protocol.ctl protocol/controller.Controller
func (c *client) connRead(ctx context.Context) {
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)

	defer ctx.Cancel(fmt.Errorf("connection read finished"))

	for {
		select {
		case <-ctx.Done():
			return
		case <-protCtl.Finished():
			return
		case mes := <-protCtl.MessageGet():
			if mes == nil {
				return
			}

			ctx.Child("mes.process", c.connMesProcess).
				ValueSet("mes", mes).Go()
		case req := <-protCtl.RequestGet():
			if req == nil {
				return
			}

			panic(fmt.Errorf("unsupported request: %T", req))
		}
	}
}

func (c *client) taskNew(queueId uint64, stateId uint64, uuid string, parentUUID string, status string) *task {
	c.task.mu.Lock()
	defer c.task.mu.Unlock()

	t := taskNew()

	it := c.task.list[stateId]

	switch ot := it.(type) {
	case *mes.SC_TaskCancel_ms:
		t.cancel(fmt.Errorf(ot.Reason))
		delete(c.task.list, stateId)
	case *task:
		if ot.queueId != 0 && queueId != 0 && ot.queueId != queueId ||
			ot.uuid != uuid ||
			ot.parentUUID != parentUUID {

			panic(fmt.Errorf("Old task with same state id (%d) but different datas: queueId old(%d) new(%d), uuid old(%s) new(%s) parentUUD old(%s) new(%s) ",
				stateId, ot.queueId, queueId, ot.uuid, uuid, ot.parentUUID, parentUUID))
		}

		t = ot

		if ot.stateId > stateId {
			return ot
		}

	default:
		c.task.list[stateId] = t
	}

	t.ctx = c.ctx.worker
	t.protocol.ctl = c.protocol.ctl
	t.statusSubscribe.do = c.taskStatusSubscribe

	t.queueId = queueId
	t.stateId = stateId
	t.uuid = uuid
	t.parentUUID = parentUUID
	t.status = status

	return t
}

func (c *client) queueTasksSubscribe(subscribeId uint64, queueId uint64, ch chan Task) {
	c.queue.mu.Lock()
	defer c.queue.mu.Unlock()

	c.queue.tasksSubscribe[subscribeId] = &subInfo{queueId, ch}
}

func (c *client) taskStatusSubscribe(subscribeId uint64, queueId uint64) (ch chan Task) {
	c.task.mu.Lock()
	defer c.task.mu.Unlock()

	sii := c.task.statusSubscribe[subscribeId]

	switch si := sii.(type) {
	case chan struct{}:
		close(si)
	case *subInfo:
		return si.ch
	}

	si := &subInfo{queueId, make(chan Task)}
	c.task.statusSubscribe[subscribeId] = si

	return si.ch
}

// ctx should contain vars:
//   mes *protocol.Message
func (c *client) connMesProcess(ctx context.Context) {
	inMes := ctx.Value("mes").(protocol.Message)

	switch m := inMes.(type) {
	default:
		log.Printf("TMClient: unsupported message: %T\n", inMes)

	case *controller.ErrorReadFail:
		log.Printf("TMClient: Read fail. %s\n", m.Orig)

	case *controller.ErrorUnhandledResponse:
		log.Printf("TMClient: Unhandled response. Id=%d", m.Mes.GetResponseId())

	case *mes.SC_QueueSubscribeTask_ms:
		c.queue.mu.Lock()
		si, ok := c.queue.tasksSubscribe[m.SubscribeId]
		c.queue.mu.Unlock()

		if !ok {
			log.Printf("TMClient: not found tasks subscribe. subscribeId=%d. Task rejected\n", m.SubscribeId)
			c.taskNew(0, m.Info.StateId, m.Info.UUID, m.Info.ParentUUID, m.Info.Status).Reject()
			return
		}

		t := c.taskNew(si.queueId, m.Info.StateId, m.Info.UUID, m.Info.ParentUUID, m.Info.Status)

		select {
		case <-ctx.Done():
			t.Reject()
			return
		case si.ch <- t:
		}

	case *mes.SC_TaskCancel_ms:
		c.task.mu.Lock()
		defer c.task.mu.Unlock()

		it, _ := c.task.list[m.StateId]

		if it == nil {
			c.task.list[m.StateId] = m
			return
		}

		it.(*task).cancel(fmt.Errorf(m.Reason))
		delete(c.task.list, m.StateId)

	case *mes.SC_TaskStatus_ms:
		var sii interface{}

	wait:
		for {
			c.task.mu.Lock()

			sii = c.task.statusSubscribe[m.SubscribeId]

			switch sii.(type) {
			case *subInfo:
				c.task.mu.Unlock()
				break wait
			case nil:
				sii = make(chan struct{})
				c.task.statusSubscribe[m.SubscribeId] = sii
			}

			c.task.mu.Unlock()

			ch := sii.(chan struct{})

			select {
			case <-ctx.Done():
				return
			case <-ch:
			}
		}

		si := sii.(*subInfo)

		var t Task // FUCKING GO!!!!! nil obj != nil interface

		if m.Info != nil {
			tt := c.taskNew(si.queueId, m.Info.StateId, m.Info.UUID, m.Info.ParentUUID, m.Info.Status)

			if tt.stateId != m.Info.StateId {
				return
			}

			t = tt
		}

		select {
		case <-ctx.Done():
			return
		case si.ch <- t:
		}
	}
}
