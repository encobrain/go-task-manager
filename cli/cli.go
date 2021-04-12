package cli

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

type Config interface {
	LoadFromDir(dirPath string) (err error)
}

type Command interface {
	flags.Commander

	Name() string
	Description() string
	LongDescription() string
}

type preOptions struct {
	configPath string `short:"c"   long:"config.path"   env:"CONFIG_PATH"   description:"A path with configuration files"   default:"config-dev"`

	help  bool `short:"h"   long:"help"   hidden:"true"`
	help2 bool `short:"?"   hidden:"true"`
}

type options struct {
	preOptions

	config Config
}

type cli struct {
	preOptions
	options

	preParser, parser *flags.Parser
}

func New() *cli {
	cli := new(cli)
	cli.preParser = flags.NewParser(&cli.preOptions, flags.IgnoreUnknown|flags.PassAfterNonOption|flags.PassDoubleDash)
	cli.parser = flags.NewParser(&cli.options, flags.Default)

	return cli
}

func (c *cli) SetConfig(config Config) {
	c.config = config
}

func (c *cli) AddCommand(cmd Command) error {
	_, err := c.parser.AddCommand(cmd.Name(), cmd.Description(), cmd.LongDescription(), cmd)
	return err
}

func (c *cli) Run() {
	_, err := c.preParser.Parse()

	switch {
	case err != nil:
		log.Printf("Parse cli options fail. %s\n", err)
		os.Exit(1)

	case !c.preOptions.help && !c.preOptions.help2:
		err := c.options.config.LoadFromDir(c.configPath)

		if err != nil {
			log.Printf("Load config fail. %s\n", err)
			os.Exit(1)
		}
	}

	if _, err := c.parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			} else {
				log.Printf("Unknown error. %s", flagsErr)
				os.Exit(1)
			}
		} else {
			log.Printf("Parse cli options fail. %s\n", err)

			os.Exit(1)
		}
	}
}
