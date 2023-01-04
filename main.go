package main

import (
	"fmt"
	"os"

	"github.com/stirboy/jh/cmd"
	"github.com/stirboy/jh/pkg/factory"
)

func main() {
	f := factory.NewFactory()

	rootCmd := cmd.NewCmdRoot(f)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("failed to execute root command:  %s\n", err)
		os.Exit(1)
	}
}
