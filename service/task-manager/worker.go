package task_manager

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/service"
	"github.com/encobrain/go-task-manager/service/task-manager/router"
	"github.com/gorilla/websocket"
	"log"
	"runtime/debug"
)

func (s *tmService) workerPanicHandler(ctx context.Context, panicErr interface{}) {
	log.Printf("Service panic. %s\n", panicErr)
	debug.PrintStack()

	s.workerStop(fmt.Errorf("panic"))
}

func (s *tmService) workerStart(ctx context.Context) {
	ctx.PanicHandlerSet(s.workerPanicHandler)

	s.conn.serve = make(chan *websocket.Conn)
	defer close(s.conn.serve)

	ctx.Child("conn.worker", s.connWorker).Go()

	s.task.router = router.New(ctx)

	if !s.statusSet(service.StatusStarted, service.StatusStarting) {
		panic(fmt.Errorf("set started status fail"))
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker stopped. %s\n", ctx.Err())
			s.workerStop(ctx.Err())
			return
		}
	}

}

func (s *tmService) workerStop(reason error) {
	if !s.statusSet(service.StatusStopping, service.StatusStarting, service.StatusStarted) {
		return
	}

	s.ctx.worker.Cancel(reason)

	<-s.ctx.worker.ChildsFinished(true)

	s.statusSet(service.StatusStopped)
}
