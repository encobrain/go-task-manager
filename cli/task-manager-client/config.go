package main

import (
	_config "github.com/encobrain/go-task-manager/model/config"
	"log"
	"os"
)

type config struct {
	Config  _config.Config  `group:"Config options" namespace:"config"`
	Process _config.Process `group:"Process options" namespace:"process"`
	Client  _config.Client  `group:"Client options" namespace:"client"`
}

func (c *config) Init() {
	defer func() {
		err := c.Config.CatchError(recover())

		if err != nil {
			log.Printf("Load config fail. %s\n", err)
			os.Exit(1)
		}
	}()

	c.Config.Load(&c.Process)
	c.Config.Load(&c.Client)

	return
}
