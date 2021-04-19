package cmd

import (
	"github.com/encobrain/go-task-manager/cli"
)

type Stop struct {
	*Config
	cli.CmdStop
}

func (c Stop) Execute(args []string) error {
	c.Config.Init()
	return c.CmdStop.Execute(c.Config.Process.Run.PidPathfile)
}
