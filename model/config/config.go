package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/encobrain/go-task-manager/lib/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Path string `long:"path"   env:"PATH"   description:"A path with configuration files"   default:"config-dev"`
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
		if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
			log.Printf("CONFIG: file %s not found\n", pathfile)
			return
		}

		panic(fmt.Errorf("open %s fail. %s", pathfile, err))
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(conf)

	if err != nil {
		panic(fmt.Errorf("decode content from %s fail. %s", pathfile, err))
	}

	if c, ok := conf.(initer); ok {
		c.Init()
	}

	if c, ok := conf.(applier); ok {
		c.Apply()
	}
}
