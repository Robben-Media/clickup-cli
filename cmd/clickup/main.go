package main

import (
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
