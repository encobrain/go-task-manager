package mes

import (
	. "github.com/encobrain/go-task-manager/internal/protocol"
)

type CS_QueueTaskNew_rq struct {
	RqId
	QueueId    uint64
	ParentUUID string
	Status     string
	Content    []byte
}

func (CS_QueueTaskNew_rq) Code() byte {
	return 'b'
}

func (CS_QueueTaskNew_rq) New() interface{} {
	return &CS_QueueTaskNew_rq{}
}

type SC_QueueTaskNew_rs struct {
	RsId
	UUID    string // if empty - queue id invalid
	StateId uint64
}

func (SC_QueueTaskNew_rs) Code() byte {
	return 'B'
}

func (SC_QueueTaskNew_rs) New() interface{} {
	return &SC_QueueTaskNew_rs{}
}

/////////////////////////////////////

type TaskInfo struct {
	StateId    uint64
	UUID       string
	ParentUUID string
	Status     string
}

/////////////////////////////////////

type CS_QueueTaskGet_rq struct {
	RqId
	QueueId uint64
	UUID    string
}

func (CS_QueueTaskGet_rq) Code() byte {
	return 'c'
}

func (CS_QueueTaskGet_rq) New() interface{} {
	return &CS_QueueTaskGet_rq{}
}

type SC_QueueTaskGet_rs struct {
	RsId
	Info *TaskInfo `json:",omitempty"` // if nil - task not exists
}

func (SC_QueueTaskGet_rs) Code() byte {
	return 'C'
}

func (SC_QueueTaskGet_rs) New() interface{} {
	return &SC_QueueTaskGet_rs{}
}

//////////////////////////////////////

type CS_QueueTasksSubscribe_rq struct {
	RqId
	QueueId    uint64
	ParentUUID string
}

func (CS_QueueTasksSubscribe_rq) Code() byte {
	return 'e'
}

func (CS_QueueTasksSubscribe_rq) New() interface{} {
	return &CS_QueueTasksSubscribe_rq{}
}

type SC_QueueTasksSubscribe_rs struct {
	RsId
	SubscribeId *uint64 // if nil - queue not exists. Individual for each connection
}

func (SC_QueueTasksSubscribe_rs) Code() byte {
	return 'E'
}

func (SC_QueueTasksSubscribe_rs) New() interface{} {
	return &SC_QueueTasksSubscribe_rs{}
}

type SC_QueueSubscribeTask_ms struct {
	SubscribeId uint64
	Info        TaskInfo
}

func (SC_QueueSubscribeTask_ms) Code() byte {
	return 'F'
}

func (SC_QueueSubscribeTask_ms) New() interface{} {
	return &SC_QueueSubscribeTask_ms{}
}

////////////////////////////////////

type CS_QueueTasksUnsubscribe_rq struct {
	RqId
	SubscribeId uint64
}

func (CS_QueueTasksUnsubscribe_rq) Code() byte {
	return 'n'
}

func (CS_QueueTasksUnsubscribe_rq) New() interface{} {
	return &CS_QueueTasksUnsubscribe_rq{}
}

type SC_QueueTasksUnsubscribe_rs struct {
	RsId
}

func (SC_QueueTasksUnsubscribe_rs) Code() byte {
	return 'N'
}

func (SC_QueueTasksUnsubscribe_rs) New() interface{} {
	return &SC_QueueSubscribeTask_ms{}
}

////////////////////////////////////

type CS_QueueTasksGet_rq struct {
	RqId
	QueueId    uint64
	ParentUUID string
}

func (CS_QueueTasksGet_rq) Code() byte {
	return 'd'
}

func (CS_QueueTasksGet_rq) New() interface{} {
	return &CS_QueueTasksGet_rq{}
}

type SC_QueueTasksGet_rs struct {
	RsId
	Tasks []TaskInfo
}

func (SC_QueueTasksGet_rs) Code() byte {
	return 'D'
}

func (SC_QueueTasksGet_rs) New() interface{} {
	return &SC_QueueTasksGet_rs{}
}
