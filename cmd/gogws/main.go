package main

import (
	"fmt"
	"os"

	"gogws/internal/commands"
)

var version = "1.0.0"

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
