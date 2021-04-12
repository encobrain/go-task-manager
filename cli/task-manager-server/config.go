package main

import (
	_config "github.com/encobrain/go-task-manager/model/config"
	"log"
	"os"
)

type config struct {
	_config.Config  `group:"Config options" namespace:"config"`
	_config.Process `group:"Process options" namespace:"process"`
	_config.Server  `group:"Server options" namespace:"server"`
}

func (c *config) Init() {
	defer func() {
		err := c.CatchError(recover())

		if err != nil {
			log.Printf("Load config fail. %s\n", err)
			os.Exit(1)
		}
	}()

	c.Load(&c.Process)
	c.Load(&c.Server)

	return
}
