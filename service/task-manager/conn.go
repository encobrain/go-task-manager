package task_manager

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/gorilla/websocket"
	"log"
)

func (s *tmService) ConnServe(conn *websocket.Conn) (err error) {
	defer func() {
		e := recover()

		if e != nil {
			e = fmt.Errorf("not started")
		}
	}()

	s.conn.serve <- conn

	return
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/QueueManager
func (s *tmService) connWorker(ctx context.Context) {
	defer close(s.conn.serve)

loop:
	for {
		select {
		case <-ctx.Done():
			log.Printf("Conn worker stopped. %s\n", ctx.Err())
			break loop
		case conn := <-s.conn.serve:
			if conn == nil {
				continue
			}

			ctx := ctx.Child("conn.serve", s.connServe)
			ctx.ValueSet("conn", conn)
		}
	}
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/QueueManager
//
//   conn *github.com/gorilla/websocket.Conn
func (s *tmService) connServe(ctx context.Context) {
	conn := ctx.Value("conn").(*websocket.Conn)
	defer conn.Close()

	ctx.PanicHandlerSet(func(ctx context.Context, panicVal interface{}) {
		if panicVal != true {
			log.Printf("Connection serve panic. %s\n", panicVal)
		}

		ctx.Cancel(fmt.Errorf("panic"))
	})

	protCtl := controller.New(protocol.Codes, conn)

	ctx.ValueSet("protocol.ctl", protCtl)
	ctx.ValueSet("task.state", newTaskState())
	ctx.ValueSet("queue.subscribe.state", newQueueSubscribeState())

	for {
		select {
		case <-ctx.Done():
			return
		case mes, ok := <-protCtl.MessageGet():
			if !ok {
				log.Printf("Protocol finished work\n")
				return
			}

			ctx := ctx.Child("mes.process", s.mesProcess)
			ctx.ValueSet("mes", mes)
		case req, ok := <-protCtl.RequestGet():
			if !ok {
				log.Printf("Protocol finished work\n")
				return
			}

			ctx := ctx.Child("req.process", s.reqProcess)
			ctx.ValueSet("req", req)
		}
	}
}