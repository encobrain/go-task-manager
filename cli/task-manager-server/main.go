package main

import (
	"github.com/encobrain/go-task-manager/cli"
	"github.com/encobrain/go-task-manager/cli/task-manager-server/cmd"
	"log"
	"os"
)

func main() {
	config := &cmd.Config{}
	cli := cli.New(config)

	err := cli.AddCommand(&cmd.Start{Config: config})

	if err != nil {
		log.Printf("Add command start fail. %s", err)
		os.Exit(1)
	}

	err = cli.AddCommand(&cmd.Stop{Config: config})

	if err != nil {
		log.Printf("Add command stop fail. %s", err)
		os.Exit(1)
	}

	cli.Run()
}
