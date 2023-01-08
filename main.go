package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/cmd"
	"github.com/stirboy/jh/pkg/cmd/gem"
	"github.com/stirboy/jh/pkg/cmd/jira/auth"
	"github.com/stirboy/jh/pkg/factory"
	"github.com/stirboy/jh/pkg/utils"
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
		fmt.Printf("error occured: %v\n", err)
		os.Exit(1)
	}

	if showGem() {
		gem.GetRandomGem()
	}
}

func showGem() bool {
	// only jh is called
	if len(os.Args) == 1 {
		return false
	}
	// help function is called
	if len(os.Args) > 1 && (utils.Contains(os.Args, "--help") || utils.Contains(os.Args, "-h")) {
		return false
	}

	return true
}

func IsUserCancellation(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}
