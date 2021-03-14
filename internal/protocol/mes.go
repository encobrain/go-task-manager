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
	0: &mes.CS_ClientGetQueue_rq{},
	1: &mes.CS_QueueTaskNew_rq{},
	2: &mes.CS_QueueTaskGet_rq{},
	3: &mes.CS_QueueTasksGet_rq{},
	4: &mes.CS_QueueTasksSubscribe_rq{},
	5: &mes.CS_TaskStatusSet_rq{},
	6: &mes.CS_TaskStatusSubscribe_rq{},
	7: &mes.CS_TaskContent_rq{},
	8: &mes.CS_TaskReject_rq{},
	9: &mes.CS_TaskRemove_rq{},

	100: &mes.SC_ClientGetQueue_rs{},
	101: &mes.SC_QueueTaskNew_rs{},
	102: &mes.SC_QueueTaskGet_rs{},
	103: &mes.SC_QueueTasksGet_rs{},
	104: &mes.SC_QueueTasksSubscribe_rs{},
	105: &mes.SC_QueueSubscribeTask_ms{},
	106: &mes.SC_TaskStatusSet_rs{},
	107: &mes.SC_TaskStatusSubscribe_rs{},
	108: &mes.SC_TaskStatus_ms{},
	109: &mes.SC_TaskContent_rs{},
	110: &mes.SC_TaskReject_rs{},
	111: &mes.SC_TaskRemove_rs{},
	112: &mes.SC_TaskCancel_ms{},
}
