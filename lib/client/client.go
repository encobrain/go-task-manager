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

type client struct {
	mu sync.Mutex

	conf *config.Client

	ctx struct {
		glob   context.Context
		worker context.Context
	}

	queue sync.Map // map[name]*queue
	task  sync.Map // map[stateId]*task

	protocol struct {
		ctl chan controller.Controller // if closed - client stopped
	}
}

func (c *client) GetQueue(name string) (queue <-chan Queue) {
	ch := make(chan Queue, 1)

	c.ctx.glob.Child("queue.get", func(ctx context.Context) {
		defer close(ch)

		queue, ok := c.queue.Load(name)

		if ok {
			ch <- queue.(Queue)
			return
		}

		for {
			protCtl := <-c.protocol.ctl

			if protCtl == nil {
				return
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

				q := c.queueNew(rs.QueueId)

				queue, _ := c.queue.LoadOrStore(name, q)

				ch <- queue.(Queue)
				return
			}
		}
	})

	return ch
}

func (c *client) queueNew(id uint64) *queue {
	q := newQueue()
	q.id = id
	q.task = &c.task
	q.protocol.ctl = c.protocol.ctl
	q.ctx = c.ctx.worker

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
		ctx.Cancel(fmt.Errorf("panic"))
	})

	ctx.Child("conn.worker", c.connWorker).Go()

	defer close(c.protocol.ctl)

	<-ctx.Done()

	log.Printf("TMClient: worker stopped. %s\n", ctx.Err())
}

func (c *client) connWorker(ctx context.Context) {

	for {
		addr := fmt.Sprintf("%s:%d", c.conf.Connect.Host, c.conf.Connect.Port)
		u := url.URL{Scheme: c.conf.Connect.Scheme, Host: addr, Path: c.conf.Connect.Path}

		log.Printf("TMClient: connecting to %s...\n", u.String())

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)

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

		protCtl := controller.New(mes.Codes, conn)

		ctx.Child("conn.read", c.connRead).
			ValueSet("protocol.ctl", protCtl).Go()

	wait:
		for {
			select {
			case <-ctx.Done():
				return
			case <-protCtl.Finished():
				break wait
			case c.protocol.ctl <- protCtl:
			}
		}
	}
}

// ctx should contain vars:
//   protocol.ctl protocol/controller.Controller
func (c *client) connRead(ctx context.Context) {
	protCtl := ctx.Value("protocol.ctl").(controller.Controller)

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
			panic(fmt.Errorf("unsupported request: %T", req))
		}
	}
}

// ctx should contain vars:
//   mes *protocol.Message
func (c *client) connMesProcess(ctx context.Context) {
	inMes := ctx.Value("mes").(protocol.Message)

	switch inMes.(type) {
	default:
		panic(fmt.Errorf("unsupported message: %T", inMes))
	case *mes.SC_QueueSubscribeTask_ms:
	case *mes.SC_TaskCancel_ms:
	case *mes.SC_TaskStatus_ms:
	}
}
