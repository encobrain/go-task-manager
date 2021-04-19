package cmd

import (
	"github.com/encobrain/go-task-manager/cli"
)

type Stop struct {
	cli.CmdStop
}

func (c Stop) Execute(args []string) error {
	Config.Init()
	return c.CmdStop.Execute(Config.Process.Run.PidPathfile)
}
