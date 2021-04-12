package main

import "github.com/encobrain/go-task-manager/cli"

type CmdStop struct {
	cli.CmdStop
}

func (c CmdStop) Execute(args []string) error {
	conf.Init()
	return c.CmdStop.Execute(conf.Process.Run.PidPathfile)
}
