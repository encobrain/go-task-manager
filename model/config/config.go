package config

import (
	"fmt"
	"os"

	"github.com/encobrain/go-task-manager/lib/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Path string `long:"path"   env:"CONFIG_PATH"   description:"A path with configuration files"   default:"config-dev"`
}

type pathfiler interface {
	Pathfile() string
}

type initer interface {
	Init()
}

type applier interface {
	Apply()
}

func (c *Config) CatchError(e interface{}) (err error) {
	if e == nil {
		return
	}

	switch e := e.(type) {
	case error:
		err = e
	default:
		err = fmt.Errorf("%s", e)
	}

	return
}

func (c *Config) Load(conf pathfiler) {
	pathfile := conf.Pathfile()

	f, err := os.Open(filepath.Resolve(true, c.Path, pathfile))

	if err != nil {
		panic(err)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(conf)

	if err != nil {
		panic(err)
	}

	if c, ok := conf.(initer); ok {
		c.Init()
	}

	if c, ok := conf.(applier); ok {
		c.Apply()
	}
}
