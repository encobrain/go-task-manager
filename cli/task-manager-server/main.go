package main

import (
	"github.com/encobrain/go-task-manager/cli"
	"log"
	"os"
)

var conf = &config{}

func main() {
	cli := cli.New(conf)

	err := cli.AddCommand(&CmdStart{})

	if err != nil {
		log.Printf("Add command start fail. %s", err)
		os.Exit(1)
	}

	err = cli.AddCommand(&CmdStop{})

	if err != nil {
		log.Printf("Add command stop fail. %s", err)
		os.Exit(1)
	}

	cli.Run()
}
