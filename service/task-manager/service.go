package task_manager

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	confService "github.com/encobrain/go-task-manager/model/config/service"
	"github.com/encobrain/go-task-manager/service"
	"github.com/encobrain/go-task-manager/service/task-manager/router"
	"github.com/gorilla/websocket"

	"sync"
)

type Service interface {
	service.Service

	// ctx should contain vars:
	//   storage.queue.manager lib/storage/queue.Manager
	Start()
	ConnServe(conn *websocket.Conn) error
}

// ctx should contain vars:
//   config *model/config/service.TaskManager
func New(ctx context.Context) Service {
	s := &tmService{}
	s.ctx.glob = ctx
	s.conf = ctx.Value("config").(*confService.TaskManager)

	s.status.waiters = map[service.Status][]chan struct{}{}
	s.status.v = service.StatusStopped

	s.conn.serve = make(chan *websocket.Conn)
	close(s.conn.serve)

	return s
}

type tmService struct {
	status struct {
		v       service.Status
		lock    sync.RWMutex
		waiters map[service.Status][]chan struct{}
	}

	conf *confService.TaskManager

	ctx struct {
		glob   context.Context
		worker context.Context
	}

	task struct {
		router *router.Router
	}

	conn struct {
		serve chan *websocket.Conn
	}
}

func (s *tmService) Start() {
	if !s.statusSet(service.StatusStarting, service.StatusStopped) {
		return
	}

	s.ctx.worker = s.ctx.glob.Child("worker", s.workerStart).Go()
}

func (s *tmService) Stop() {
	if !s.statusSet(service.StatusStopping, service.StatusStarting, service.StatusStarted) {
		return
	}

	s.ctx.worker.Cancel(fmt.Errorf("service stopped"))
}
