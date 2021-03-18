package task_manager

import (
	"encoding/json"
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol"
	"github.com/encobrain/go-task-manager/internal/protocol/mes"
)

// ctx should contain vars:
//   task.state *taskState
//   protocol.ctl protocol/controller.Controller
//   storage.queue.manager lib/storage/QueueManager
//   queue.subscribe.state *queueSubscribeState
//
//   req protocol.Request
func (s *tmService) reqProcess(ctx context.Context) {
	ireq := ctx.Value("req").(protocol.Request)

	switch req := ireq.(type) {
	default:
		bytes, _ := json.Marshal(req)
		panic(fmt.Errorf("unsupported request. %T %s", req, string(bytes)))
	case *mes.CS_ClientGetQueue_rq:
		s.queueGet(ctx)
	case *mes.CS_QueueTaskNew_rq:
		s.queueTaskNew(ctx)
	case *mes.CS_QueueTaskGet_rq:
		s.queueTaskGet(ctx)
	case *mes.CS_QueueTasksSubscribe_rq:
		s.queueTaskSubscribe(ctx)
	}
}
