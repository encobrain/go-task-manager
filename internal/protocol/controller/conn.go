package controller

import (
	"encoding/json"
	"fmt"
	"github.com/encobrain/go-task-manager/internal/protocol"
	"github.com/gorilla/websocket"
	"log"
)

func (c *controller) connIncomingReq(req protocol.Request) {
	j, _ := json.Marshal(req)
	log.Printf("TMProtocol<%s>: incoming req %T %s\n", c.conn.RemoteAddr(), req, j)

	defer func() { recover() }()
	c.incoming.reqs <- req
}

func (c *controller) connIncomingMes(mes protocol.Message) {
	j, _ := json.Marshal(mes)
	log.Printf("TMProtocol<%s>: incoming mes %T %s\n", c.conn.RemoteAddr(), mes, j)

	defer func() { recover() }()
	c.incoming.mess <- mes
}

func (c *controller) connIncomingRes(ch chan protocol.Response, res protocol.Response) {
	j, _ := json.Marshal(res)
	log.Printf("TMProtocol<%s>: incoming res %T %s\n", c.conn.RemoteAddr(), res, j)

	defer func() {
		err := recover()
		if err == nil {
			return
		}

		log.Printf("TMProtocol<%s>: send incoming res %T %s fail. Channel closed\n", c.conn.RemoteAddr(), res, j)
	}()
	ch <- res
	close(ch)
}

func (c *controller) connRead() {
	defer c.connStop()

	for {
		mt, bytes, err := c.conn.ReadMessage()

		if err != nil {
			err := &ErrorReadFail{Orig: err}

			go c.connIncomingMes(err)
			break
		}

		if mt != websocket.TextMessage {
			err := &ErrorReadFail{Orig: fmt.Errorf("invalid message format. type=%v", mt)}
			go c.connIncomingMes(err)
			break
		}

		if len(bytes) < 1 {
			err := &ErrorReadFail{Orig: fmt.Errorf("zero message size")}
			go c.connIncomingMes(err)
			break
		}

		code := bytes[0]

		mesStruct, ok := c.codeMes[code]

		if !ok {
			err := &ErrorReadFail{Orig: fmt.Errorf("unknown message type. type=%d", code)}
			go c.connIncomingMes(err)
			break
		}

		mes := mesStruct.New()

		err = json.Unmarshal(bytes[1:], mes)

		if err != nil {
			err := &ErrorReadFail{Orig: fmt.Errorf("unmarshal message fail. %s", err)}
			go c.connIncomingMes(err)
			break
		}

		switch m := mes.(type) {
		case protocol.Request:
			go c.connIncomingReq(m)

		case protocol.Response:
			id := m.GetResponseId()
			resCh, ok := c.res.list.Load(id)

			if !ok {
				err := &ErrorUnhandledResponse{Mes: m}
				go c.connIncomingMes(err)
				continue
			}

			c.res.list.Delete(id)

			go c.connIncomingRes(resCh.(chan protocol.Response), m)

		default:
			go c.connIncomingMes(m.(protocol.Message))
		}
	}
}

func (c *controller) connStop() {
	close(c.incoming.mess)
	close(c.incoming.reqs)
	close(c.finished)
	c.conn.Close()

	c.res.list.Range(func(key, value interface{}) (ok bool) {
		ok = true
		ch := value.(chan protocol.Response)

		defer func() { recover() }()
		close(ch)
		return
	})

	log.Printf("TMProtocol<%s>: process stopped", c.conn.RemoteAddr())
}
