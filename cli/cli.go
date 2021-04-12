package cli

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

type Command interface {
	flags.Commander

	Name() string
	Description() string
	LongDescription() string
}

type cli struct {
	parser *flags.Parser
}

func New(options interface{}) *cli {
	cli := new(cli)
	cli.parser = flags.NewParser(options, flags.Default)

	return cli
}

func (c *cli) AddCommand(cmd Command) error {
	_, err := c.parser.AddCommand(cmd.Name(), cmd.Description(), cmd.LongDescription(), cmd)
	return err
}

func (c *cli) Run() {
	if _, err := c.parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp || flagsErr.Type == flags.ErrCommandRequired {
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
