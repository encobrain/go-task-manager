package cli

import (
	"log"
	"os"
	"syscall"
	"time"
)

type CmdStop struct{}

func (CmdStop) Name() string {
	return "stop"
}

func (CmdStop) Description() string {
	return "Stop the service"
}

func (CmdStop) LongDescription() string {
	return "The Stop command stops the service if it not stopped"
}

func (CmdStop) Execute(pidPathfile string) error {
	process, err := ProcessGet(pidPathfile)

	if err != nil {
		return err
	}

	if process == nil {
		log.Printf("Service not started\n")
		return nil
	}

	err = process.Signal(os.Interrupt)

	if err != nil {
		log.Printf("Service not started\n")
		return nil
	}

	log.Printf("Stopping service...\n")

	for {
		err = process.Signal(syscall.Signal(0))

		if err != nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	log.Printf("Service stopped\n")

	return nil
}
