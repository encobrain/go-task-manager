package task_manager

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/service"
	"github.com/encobrain/go-task-manager/service/task-manager/router"
	"github.com/gorilla/websocket"
	"log"
)

func (s *tmService) workerPanicHandler(ctx context.Context, panicErr interface{}) {
	log.Printf("Service panic. %s\n", panicErr)

	ctx.Cancel(fmt.Errorf("service panic. %s", panicErr))

	s.workerStop()
}

func (s *tmService) workerStart(ctx context.Context) {
	ctx.PanicHandlerSet(s.workerPanicHandler)

	s.conn.serve = make(chan *websocket.Conn)

	s.task.router = router.New(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker stopped. %s\n", ctx.Err())
			s.workerStop()
			return
		}
	}

}

func (s *tmService) workerStop() {
	if !s.statusSet(service.StatusStopping, service.StatusStarting, service.StatusStarted) {
		return
	}

	<-s.ctx.worker.ChildsFinished(true)

	s.statusSet(service.StatusStopped)
}