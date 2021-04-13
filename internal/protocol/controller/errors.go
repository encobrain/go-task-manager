package controller

import (
	"fmt"
	"github.com/encobrain/go-task-manager/internal/protocol"
)

type ErrorReadFail struct {
	Orig error
}

func (e *ErrorReadFail) Error() string {
	return fmt.Sprintf("read fail. %s", e.Orig)
}

func (e *ErrorReadFail) New() interface{} {
	return &ErrorReadFail{}
}

func (e ErrorReadFail) Code() byte {
	return 0
}

type ErrorUnhandledResponse struct {
	Mes protocol.Response
}

func (e *ErrorUnhandledResponse) Error() string {
	return fmt.Sprintf("unhandled response. id=%d", e.Mes.GetResponseId())
}

func (e *ErrorUnhandledResponse) New() interface{} {
	return &ErrorUnhandledResponse{}
}

func (e ErrorUnhandledResponse) Code() byte {
	return 0
}
