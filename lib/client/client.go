package client

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"github.com/encobrain/go-task-manager/model/config"
	"log"
	"sync"
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
	c := &client{}
	c.ctx.glob = ctx
	c.protocol.ctl = make(chan controller.Controller)
	close(c.protocol.ctl)

	return c
}

type client struct {
	mu sync.Mutex

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

				q := newQueue(c.ctx.glob, rs.QueueId)
				q.protocol.ctl = c.protocol.ctl
				q.task = &c.task

				queue, _ := c.queue.LoadOrStore(name, q)

				ch <- queue.(Queue)
				return
			}
		}
	})

	return ch
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

}
