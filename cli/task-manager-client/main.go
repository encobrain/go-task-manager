package main

import (
	"github.com/encobrain/go-task-manager/cli"
	"log"
	"os"
)

var conf = &config{}

func main() {
	cli := cli.New(conf)

	err := cli.AddCommand(&CmdTaskAdd{})

	if err != nil {
		log.Printf("Add command task_add fail. %s", err)
		os.Exit(1)
	}

	cli.Run()
}
