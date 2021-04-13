package mes

import (
	. "github.com/encobrain/go-task-manager/internal/protocol"
)

type CS_TaskStatusSubscribe_rq struct {
	RqId
	QueueId uint64
	UUID    string
}

func (CS_TaskStatusSubscribe_rq) Code() byte {
	return 'g'
}

func (CS_TaskStatusSubscribe_rq) New() interface{} {
	return &CS_TaskStatusSubscribe_rq{}
}

type SC_TaskStatusSubscribe_rs struct {
	RsId
	SubscribeId *uint64 // if nil - queue or task not exists. Individual for each connection
}

func (SC_TaskStatusSubscribe_rs) Code() byte {
	return 'H'
}

func (SC_TaskStatusSubscribe_rs) New() interface{} {
	return &SC_TaskStatusSubscribe_rs{}
}

type SC_TaskStatus_ms struct {
	SubscribeId uint64
	Info        *TaskInfo // if nil - task not exists
}

func (SC_TaskStatus_ms) Code() byte {
	return 'I'
}

func (SC_TaskStatus_ms) New() interface{} {
	return &SC_TaskStatus_ms{}
}

/////////////////////////////////

type CS_TaskContent_rq struct {
	RqId
	StateId uint64
}

func (CS_TaskContent_rq) Code() byte {
	return 'h'
}

func (CS_TaskContent_rq) New() interface{} {
	return &CS_TaskContent_rq{}
}

type SC_TaskContent_rs struct {
	RsId
	Content *[]byte `json:",omitempty"` // if nil - task canceled
}

func (SC_TaskContent_rs) Code() byte {
	return 'J'
}

func (SC_TaskContent_rs) New() interface{} {
	return &SC_TaskContent_rs{}
}

//////////////////////////////////

type CS_TaskStatusSet_rq struct {
	RqId
	QueueId uint64
	StateId uint64
	Status  string
	Content []byte
}

func (CS_TaskStatusSet_rq) Code() byte {
	return 'f'
}

func (CS_TaskStatusSet_rq) New() interface{} {
	return &CS_TaskStatusSet_rq{}
}

type SC_TaskStatusSet_rs struct {
	RsId
	StateId *uint64 // if nil - task canceled/not exists or queue invalid
}

func (SC_TaskStatusSet_rs) Code() byte {
	return 'G'
}

func (SC_TaskStatusSet_rs) New() interface{} {
	return &SC_TaskStatusSet_rs{}
}

/////////////////////////////////

type CS_TaskRemove_rq struct {
	RqId
	QueueId uint64
	StateId uint64
}

func (CS_TaskRemove_rq) Code() byte {
	return 'j'
}

func (CS_TaskRemove_rq) New() interface{} {
	return &CS_TaskRemove_rq{}
}

type SC_TaskRemove_rs struct {
	RsId
	Ok bool // if false - task canceled/not exists or queue invalid
}

func (SC_TaskRemove_rs) Code() byte {
	return 'L'
}

func (SC_TaskRemove_rs) New() interface{} {
	return &SC_TaskRemove_rs{}
}

///////////////////////////////////

type CS_TaskReject_rq struct {
	RqId
	StateId uint64
}

func (CS_TaskReject_rq) Code() byte {
	return 'i'
}

func (CS_TaskReject_rq) New() interface{} {
	return &CS_TaskReject_rq{}
}

type SC_TaskReject_rs struct {
	RsId
}

func (SC_TaskReject_rs) Code() byte {
	return 'K'
}

func (SC_TaskReject_rs) New() interface{} {
	return &SC_TaskReject_rs{}
}

////////////////////////////////////

type SC_TaskCancel_ms struct {
	StateId uint64
	Reason  string
}

func (SC_TaskCancel_ms) Code() byte {
	return 'M'
}

func (SC_TaskCancel_ms) New() interface{} {
	return &SC_TaskCancel_ms{}
}
