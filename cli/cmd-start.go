package cli

import "log"

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
		return
	}

	if is {
		log.Printf("Service already started\n")
	} else {
		log.Printf("Starting service...")

		err = ProcessSavePid(pidPathfile)

		if err == nil {
			success = true
		}
	}

	return
}
