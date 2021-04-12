package main

import (
	"errors"
	"github.com/encobrain/go-task-manager/cli"
)

type CmdStart struct {
	cli.CmdStart
}

func (c CmdStart) Execute(args []string) error {
	conf.Init()
	success, err := c.CmdStart.Execute(conf.Process.Run.PidPathfile)

	if err != nil || !success {
		return err
	}

	return errors.New("not implemented")
}
