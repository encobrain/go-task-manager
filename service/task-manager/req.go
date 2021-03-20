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
//   task.status.subscribe.state *taskStatusSubscribeState
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
	case *mes.CS_QueueTasksGet_rq:
		s.queueTasksGet(ctx)

	case *mes.CS_TaskStatusSubscribe_rq:
		s.taskStatusSubscribe(ctx)
	case *mes.CS_TaskContent_rq:
		s.taskContent(ctx)
	case *mes.CS_TaskStatusSet_rq:
		s.taskStatusSet(ctx)
	case *mes.CS_TaskRemove_rq:
		s.taskRemove(ctx)
	case *mes.CS_TaskReject_rq:
		s.taskReject(ctx)
	}
}
