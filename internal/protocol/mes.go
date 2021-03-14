package protocol

import "github.com/encobrain/go-task-manager/internal/protocol/mes"

type Message interface {
	New() interface{}
}

type Request interface {
	Message
	GetRequestId() (id uint64)
	SetRequestId(id uint64)
}

type Response interface {
	Message
	SetResponseId(id uint64)
	GetResponseId() (id uint64)
}

type RqId uint64

func (i RqId) GetRequestId() (id uint64) {
	return uint64(i)
}

func (i *RqId) SetRequestId(id uint64) {
	*i = RqId(id)
}

type RsId uint64

func (i *RsId) SetResponseId(id uint64) {
	*i = RsId(id)
}

func (i RsId) GetResponseId() (id uint64) {
	return uint64(i)
}

var Codes = map[byte]Message{
	'a': &mes.CS_ClientGetQueue_rq{},
	'b': &mes.CS_QueueTaskNew_rq{},
	'c': &mes.CS_QueueTaskGet_rq{},
	'd': &mes.CS_QueueTasksGet_rq{},
	'e': &mes.CS_QueueTasksSubscribe_rq{},
	'f': &mes.CS_TaskStatusSet_rq{},
	'g': &mes.CS_TaskStatusSubscribe_rq{},
	'h': &mes.CS_TaskContent_rq{},
	'i': &mes.CS_TaskReject_rq{},
	'j': &mes.CS_TaskRemove_rq{},

	'A': &mes.SC_ClientGetQueue_rs{},
	'B': &mes.SC_QueueTaskNew_rs{},
	'C': &mes.SC_QueueTaskGet_rs{},
	'D': &mes.SC_QueueTasksGet_rs{},
	'E': &mes.SC_QueueTasksSubscribe_rs{},
	'F': &mes.SC_QueueSubscribeTask_ms{},
	'G': &mes.SC_TaskStatusSet_rs{},
	'H': &mes.SC_TaskStatusSubscribe_rs{},
	'I': &mes.SC_TaskStatus_ms{},
	'J': &mes.SC_TaskContent_rs{},
	'K': &mes.SC_TaskReject_rs{},
	'L': &mes.SC_TaskRemove_rs{},
	'M': &mes.SC_TaskCancel_ms{},
}
