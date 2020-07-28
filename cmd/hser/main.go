package main

import (
	"github.com/masterhung0112/go_server/cmd/hser/commands"
	"os"
)

func main() {
  if err := commands.Run(os.Args[1:]); err != nil {
    os.Exit(1)
  }
}