package main

import (
	"os"

	"github.com/danhigham/emonbeat/cmd"

	// Make sure all your modules and metricsets are linked in this file
	_ "github.com/danhigham/emonbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
