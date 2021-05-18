package task_manager

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
	"github.com/gorilla/websocket"
	"log"
	"runtime/debug"
)

func (s *tmService) ConnServe(conn *websocket.Conn) (err error) {
	defer func() {
		e := recover()

		if e != nil {
			err = fmt.Errorf("not started")
		}
	}()

	if conn == nil {
		err = fmt.Errorf("conn is nil")
		return
	}

	s.conn.serve <- conn

	return
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
func (s *tmService) connWorker(ctx context.Context) {
	defer log.Printf("Conn worker stopped\n")

	for {
		select {
		case <-ctx.Done():
			return
		case conn, ok := <-s.conn.serve:
			if !ok {
				return
			}

			ctx.Child("serve", s.connServe).
				ValueSet("conn", conn).Go()
		}
	}
}

// ctx should contain vars:
//   storage.queue.manager lib/storage/queue.Manager
//
//   conn *github.com/gorilla/websocket.Conn
func (s *tmService) connServe(ctx context.Context) {
	conn := ctx.Value("conn").(*websocket.Conn)
	defer conn.Close()

	ctx.PanicHandlerSet(func(ctx context.Context, panicVal interface{}) {
		if panicVal != true {
			log.Printf("Connection serve panic. %s\n", panicVal)
			debug.PrintStack()
		}

		ctx.Cancel(fmt.Errorf("panic"))
	})

	protCtl := controller.New(mes.Messages, conn)

	ctx.ValueSet("protocol.ctl", protCtl)
	ctx.ValueSet("task.state", newTaskState(ctx))
	ctx.ValueSet("queue.subscribe.state", newQueueSubscribeState())
	ctx.ValueSet("task.status.subscribe.state", newTaskStatusSubscribeState())

	for {
		select {
		case <-ctx.Done():
			return
		case mes, ok := <-protCtl.MessageGet():
			if !ok {
				return
			}

			ctx.Child("mes.process", s.mesProcess).
				ValueSet("mes", mes).Go()
		case req, ok := <-protCtl.RequestGet():
			if !ok {
				return
			}

			ctx.Child("req.process", s.reqProcess).
				ValueSet("req", req).Go()
		}
	}
}
