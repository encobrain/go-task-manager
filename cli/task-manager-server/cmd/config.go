package cmd

import (
	_config "github.com/encobrain/go-task-manager/model/config"
	"github.com/encobrain/go-task-manager/model/config/lib"
	"github.com/encobrain/go-task-manager/model/config/lib/db/driver"
	"log"
	"os"
)

type config struct {
	Config          _config.Config  `group:"Config options" namespace:"config"`
	Process         _config.Process `group:"Process options" namespace:"process"`
	Server          _config.Server  `group:"Server options" namespace:"server"`
	DbDriverManager driver.Manager  `group:"DB driver manager options" namespace:"dbDriverManager"`
	Storage         lib.Storage     `group:"Storage options" namespace:"storage"`
}

var Config = &config{}

func (c *config) Init() {
	defer func() {
		err := c.Config.CatchError(recover())

		if err != nil {
			log.Printf("Load config fail. %s\n", err)
			os.Exit(1)
		}
	}()

	c.Config.Load(&c.Process)
	c.Config.Load(&c.Server)
	c.Config.Load(&c.DbDriverManager)
	c.Config.Load(&c.Storage)

	return
}
