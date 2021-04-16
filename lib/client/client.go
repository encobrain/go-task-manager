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
		list           sync.Map // map[name]*queue
		tasksSubscribe sync.Map // map[subscribeId]*subInfo
	}

	task struct {
		list            sync.Map // map[stateId]*task
		statusSubscribe sync.Map // map[subscribeId]*subInfo
	}

	protocol struct {
		ctl chan controller.Controller // if closed - client stopped
	}
}

func (c *client) GetQueue(name string) (queue <-chan Queue) {
	ch := make(chan Queue, 1)

	c.ctx.glob.Child("queue.get", func(ctx context.Context) {
		defer close(ch)

		queue, ok := c.queue.list.Load(name)

		if ok {
			ch <- queue.(Queue)
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
	q := newQueue()
	q.id = id
	q.name = name
	q.tasks.new = c.taskNew
	q.tasks.subscribe = c.queueTasksSubscribe
	q.protocol.ctl = c.protocol.ctl
	q.ctx = c.ctx.worker

	iq, _ := c.queue.list.LoadOrStore(name, q)

	return iq.(*queue)
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
	c.task.list.Range(func(stateId, itask interface{}) bool {
		task := itask.(*task)
		task.cancel(fmt.Errorf("client disconnected"))
		c.task.list.Delete(stateId)
		return true
	})

	c.task.statusSubscribe.Range(func(subId, _ interface{}) bool {
		c.task.statusSubscribe.Delete(subId)
		return true
	})

	c.queue.tasksSubscribe.Range(func(subId, _ interface{}) bool {
		c.queue.tasksSubscribe.Delete(subId)
		return true
	})
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
	t := taskNew()
	t.ctx = c.ctx.worker
	t.protocol.ctl = c.protocol.ctl
	t.statusSubscribe.do = c.taskStatusSubscribe

	t.queueId = queueId
	t.stateId = stateId
	t.uuid = uuid
	t.parentUUID = parentUUID
	t.status = status

	c.task.list.Store(stateId, t)

	return t
}

func (c *client) queueTasksSubscribe(subscribeId uint64, queueId uint64, ch chan Task) {
	c.queue.tasksSubscribe.Store(subscribeId, &subInfo{queueId, ch})
}

func (c *client) taskStatusSubscribe(subscribeId uint64, queueId uint64, ch chan Task) {
	si := &subInfo{queueId, ch}
	sii, _ := c.task.statusSubscribe.LoadOrStore(subscribeId, si)

	switch t := sii.(type) {
	case chan struct{}:
		c.task.statusSubscribe.Store(subscribeId, si)
		close(t)
	}
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
		sii, ok := c.queue.tasksSubscribe.Load(m.SubscribeId)

		if !ok {
			log.Printf("TMClient: not found tasks subscribe. subscribeId=%d. Task rejected\n", m.SubscribeId)
			c.taskNew(0, m.Info.StateId, m.Info.UUID, m.Info.ParentUUID, m.Info.Status).Reject()
			return
		}

		si := sii.(*subInfo)

		t := c.taskNew(si.queueId, m.Info.StateId, m.Info.UUID, m.Info.ParentUUID, m.Info.Status)

		select {
		case <-ctx.Done():
			t.Reject()
			return
		case si.ch <- t:
		}

	case *mes.SC_TaskCancel_ms:
		it, ok := c.task.list.Load(m.StateId)

		if !ok {
			log.Printf("TMClient: not found task for cancel. stateId=%d\n", m.StateId)
			return
		}

		it.(*task).cancel(fmt.Errorf(m.Reason))

	case *mes.SC_TaskStatus_ms:
		sii, ok := c.task.statusSubscribe.LoadOrStore(m.SubscribeId, make(chan struct{}))

		if !ok {
			log.Printf("TMClient: not found task status subscribe. subscribeId=%d. Waiting subscribe done...\n", m.SubscribeId)

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				panic(fmt.Errorf("task status subscribeId=%d not found", m.SubscribeId))
			case <-sii.(chan struct{}):
				sii, _ = c.task.statusSubscribe.Load(m.SubscribeId)
			}
		}

		si := sii.(*subInfo)

		var t *task

		if m.Info != nil {
			t = c.taskNew(si.queueId, m.Info.StateId, m.Info.UUID, m.Info.ParentUUID, m.Info.Status)
		}

		log.Printf("TMClient: Subscribe %d Sending task status %+v\n", m.SubscribeId, m.Info)

		select {
		case <-ctx.Done():
			return
		case si.ch <- t:
		}

		log.Printf("TMClient: Subscribe %d Send done\n", m.SubscribeId)
	}
}
