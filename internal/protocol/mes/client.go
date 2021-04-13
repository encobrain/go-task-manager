package mes

import (
	. "github.com/encobrain/go-task-manager/internal/protocol"
)

type CS_ClientGetQueue_rq struct {
	RqId
	Name string
}

func (CS_ClientGetQueue_rq) Code() byte {
	return 'a'
}

func (CS_ClientGetQueue_rq) New() interface{} {
	return &CS_ClientGetQueue_rq{}
}

type SC_ClientGetQueue_rs struct {
	RsId
	QueueId uint64
}

func (SC_ClientGetQueue_rs) Code() byte {
	return 'A'
}

func (SC_ClientGetQueue_rs) New() interface{} {
	return &SC_ClientGetQueue_rs{}
}
