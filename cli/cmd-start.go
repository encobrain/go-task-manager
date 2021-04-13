package cli

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type CmdStart struct {
	shutdownSignals chan os.Signal
}

func (CmdStart) Name() string {
	return "start"
}

func (CmdStart) Description() string {
	return "Start the service"
}

func (CmdStart) LongDescription() string {
	return "The Start command starts the service if it not started"
}

func (c *CmdStart) Execute(pidPathfile string) (success bool, err error) {
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
			c.shutdownSignals = make(chan os.Signal, 2)

			signal.Notify(c.shutdownSignals, syscall.SIGTERM, syscall.SIGINT)
		}
	}

	return
}

func (c *CmdStart) ShutdownSignals() <-chan os.Signal {
	return c.shutdownSignals
}
