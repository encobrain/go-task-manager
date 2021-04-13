package cli

import (
	"fmt"
	"log"
)

type CmdStart struct{}

func (CmdStart) Name() string {
	return "start"
}

func (CmdStart) Description() string {
	return "Start the service"
}

func (CmdStart) LongDescription() string {
	return "The Start command starts the service if it not started"
}

func (CmdStart) Execute(pidPathfile string) (success bool, err error) {
	is, err := ProcessIsRunned(pidPathfile)

	if err != nil {
		return false, fmt.Errorf("get process state fail. %s", err)
	}

	if is {
		log.Printf("Service already started\n")
	} else {
		log.Printf("Starting service...")

		err = ProcessSavePid(pidPathfile)

		if err != nil {
			err = fmt.Errorf("save process pid fail. %s", err)
		} else {
			success = true
		}
	}

	return
}
