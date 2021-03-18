package task_manager

import (
	"encoding/json"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/internal/protocol"
	"github.com/encobrain/go-task-manager/internal/protocol/controller"
	"log"
)

// ctx should contain vars:
//   mes protocol.Message
func (s *tmService) mesProcess(ctx context.Context) {
	imes := ctx.Value("mes").(protocol.Message)

	switch mes := imes.(type) {
	case *controller.ErrorReadFail:
		log.Printf("Read message fail. %s\n", mes.Orig)
		panic(true)
	case *controller.ErrorUnhandledResponse:
		bytes, _ := json.Marshal(mes.Mes)
		log.Printf("Readed unhandled response. %s\n", string(bytes))
	default:
		bytes, _ := json.Marshal(mes)
		log.Printf("Unsupported message. %T %s\n", mes, string(bytes))
	}
}
