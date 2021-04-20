package cmd

import (
	"github.com/encobrain/go-task-manager/cli"
)

type Stop struct {
	*Config     `no-flag:"true"`
	cli.CmdStop `no-flag:"true"`
}

func (c Stop) Execute(args []string) error {
	c.Config.Init()
	return c.CmdStop.Execute(c.Config.Process.Run.PidPathfile)
}
