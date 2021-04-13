package mes

import "github.com/encobrain/go-task-manager/internal/protocol"

var Messages = []protocol.Message{
	&CS_ClientGetQueue_rq{},
	&CS_QueueTaskNew_rq{},
	&CS_QueueTaskGet_rq{},
	&CS_QueueTasksGet_rq{},
	&CS_QueueTasksSubscribe_rq{},
	&CS_TaskStatusSet_rq{},
	&CS_TaskStatusSubscribe_rq{},
	&CS_TaskContent_rq{},
	&CS_TaskReject_rq{},
	&CS_TaskRemove_rq{},

	&SC_ClientGetQueue_rs{},
	&SC_QueueTaskNew_rs{},
	&SC_QueueTaskGet_rs{},
	&SC_QueueTasksGet_rs{},
	&SC_QueueTasksSubscribe_rs{},
	&SC_QueueSubscribeTask_ms{},
	&SC_TaskStatusSet_rs{},
	&SC_TaskStatusSubscribe_rs{},
	&SC_TaskStatus_ms{},
	&SC_TaskContent_rs{},
	&SC_TaskReject_rs{},
	&SC_TaskRemove_rs{},
	&SC_TaskCancel_ms{},
}
