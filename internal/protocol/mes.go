package protocol

type Message interface {
	New() interface{}
	Code() byte
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
