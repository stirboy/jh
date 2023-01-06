package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/cmd"
	"github.com/stirboy/jh/pkg/cmd/jira/auth"
	"github.com/stirboy/jh/pkg/factory"
)

func main() {
	f := factory.NewFactory()
	cfg, err := f.Config()
	if err != nil {
		fmt.Printf("cannot create config file %v\n", err)
		os.Exit(1)
	}

	rootCmd := cmd.NewCmdRoot(f)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// require that the user is authenticated before running most commands
		if auth.IsAuthEnabled(cmd) && !auth.CheckAuth(cfg) {
			fmt.Println("To get started with JH, please run: jh auth")
			os.Exit(1)
		}

		return nil
	}

	if err := rootCmd.Execute(); err != nil {
		if IsUserCancellation(err) {
			// ensures next shell prompt will start on a new line
			fmt.Println("")
			os.Exit(1)
		}
		fmt.Printf("error occured during the command %v\n", err)
		os.Exit(1)
	}
}

func IsUserCancellation(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}
