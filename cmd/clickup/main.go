package main

import (
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
