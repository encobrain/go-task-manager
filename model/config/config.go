package config

import (
	"fmt"
	"os"

	"github.com/encobrain/go-task-manager/lib/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct{}

type pathfiler interface {
	Pathfile() string
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

func (c *Config) Load(to pathfiler, dir string) {
	pathfile := to.Pathfile()

	f, err := os.Open(filepath.Resolve(true, dir, pathfile))

	if err != nil {
		panic(err)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(to)

	if err != nil {
		panic(err)
	}
}
