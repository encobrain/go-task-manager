package controller

import (
	"encoding/json"
	"github.com/encobrain/go-task-manager/internal/protocol"
	"github.com/gorilla/websocket"
	"sync"
	"sync/atomic"
)

type Controller interface {
	// Finished return finished channel.
	// If it closed - controller finish work with connection
	Finished() <-chan struct{}
	// Message sends message to another side
	MessageSend(mes protocol.Message) error
	// Message get returns channel with incoming messages
	// If closed - finished
	MessageGet() (mess <-chan protocol.Message)
	// Request sends request and waits response
	// Chan closed with result
	// If received nil - finished
	RequestSend(req protocol.Request) (res <-chan protocol.Response, err error)
	// RequestGet returns channel with incoming requests
	// if closed - finished
	RequestGet() (reqs <-chan protocol.Request)
	// ResponseSend sends response on request
	ResponseSend(req protocol.Request, res protocol.Response) error
}

func New(codeMes map[byte]protocol.Message, conn *websocket.Conn) Controller {
	c := &controller{
		conn:     conn,
		codeMes:  codeMes,
		finished: make(chan struct{}),
		mesCode:  make(map[protocol.Message]byte, len(codeMes)),
	}

	for code, mes := range codeMes {
		c.mesCode[mes] = code
	}

	c.incoming.mess = make(chan protocol.Message)
	c.incoming.reqs = make(chan protocol.Request)

	go c.connRead()

	return c
}

type controller struct {
	conn    *websocket.Conn
	codeMes map[byte]protocol.Message
	mesCode map[protocol.Message]byte

	incoming struct {
		mess chan protocol.Message
		reqs chan protocol.Request
	}

	finished chan struct{}

	res struct {
		nextId uint64
		list   sync.Map
	}
}

func (c *controller) MessageSend(mes protocol.Message) (err error) {
	code, ok := c.mesCode[mes]

	if !ok {
		err = ErrorUnknownMes
		return
	}

	bytes, err := json.Marshal(mes)

	if err != nil {
		return
	}

	return c.conn.WriteMessage(websocket.TextMessage, append([]byte{code}, bytes...))
}

func (c *controller) MessageGet() (mess <-chan protocol.Message) {
	return c.incoming.mess
}

func (c *controller) RequestSend(req protocol.Request) (res <-chan protocol.Response, err error) {
	id := atomic.AddUint64(&c.res.nextId, 1)
	resCh := make(chan protocol.Response, 1)

	c.res.list.Store(id, resCh)

	req.SetRequestId(id)

	err = c.MessageSend(req)

	if err != nil {
		c.res.list.Delete(id)
		close(resCh)
	}

	return
}

func (c *controller) RequestGet() (reqs <-chan protocol.Request) {
	return c.incoming.reqs
}

func (c *controller) ResponseSend(req protocol.Request, res protocol.Response) (err error) {
	res.SetResponseId(req.GetRequestId())

	return c.MessageSend(res)
}

func (c *controller) Finished() <-chan struct{} {
	return c.finished
}
