package main

import (
	"github.com/encobrain/go-task-manager/cli"
	"github.com/encobrain/go-task-manager/cli/task-manager-server/cmd"
	"log"
	"os"
)

func main() {
	cli := cli.New(cmd.Config)

	err := cli.AddCommand(&cmd.Start{})

	if err != nil {
		log.Printf("Add command start fail. %s", err)
		os.Exit(1)
	}

	err = cli.AddCommand(&cmd.Stop{})

	if err != nil {
		log.Printf("Add command stop fail. %s", err)
		os.Exit(1)
	}

	cli.Run()
}
