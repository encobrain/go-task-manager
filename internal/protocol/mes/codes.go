package mes

import "github.com/encobrain/go-task-manager/internal/protocol"

var Codes = map[byte]protocol.Message{
	'a': &CS_ClientGetQueue_rq{},
	'b': &CS_QueueTaskNew_rq{},
	'c': &CS_QueueTaskGet_rq{},
	'd': &CS_QueueTasksGet_rq{},
	'e': &CS_QueueTasksSubscribe_rq{},
	'f': &CS_TaskStatusSet_rq{},
	'g': &CS_TaskStatusSubscribe_rq{},
	'h': &CS_TaskContent_rq{},
	'i': &CS_TaskReject_rq{},
	'j': &CS_TaskRemove_rq{},

	'A': &SC_ClientGetQueue_rs{},
	'B': &SC_QueueTaskNew_rs{},
	'C': &SC_QueueTaskGet_rs{},
	'D': &SC_QueueTasksGet_rs{},
	'E': &SC_QueueTasksSubscribe_rs{},
	'F': &SC_QueueSubscribeTask_ms{},
	'G': &SC_TaskStatusSet_rs{},
	'H': &SC_TaskStatusSubscribe_rs{},
	'I': &SC_TaskStatus_ms{},
	'J': &SC_TaskContent_rs{},
	'K': &SC_TaskReject_rs{},
	'L': &SC_TaskRemove_rs{},
	'M': &SC_TaskCancel_ms{},
}
