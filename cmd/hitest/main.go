package main

import (
	"fmt"
	"os"

	"hitest/internal/core/cmd/server/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
