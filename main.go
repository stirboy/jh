package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/stirboy/jh/cmd"
	"github.com/stirboy/jh/pkg/factory"
)

func main() {
	f := factory.NewFactory()

	rootCmd := cmd.NewCmdRoot(f)
	if err := rootCmd.Execute(); err != nil {
		if IsUserCancellation(err) {
			// ensures next shell prompt will start on a new line
			fmt.Println()
			os.Exit(1)
		}
		os.Exit(1)
	}
}

func IsUserCancellation(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}
