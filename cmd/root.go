package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/cmd/jira/auth"
	jiraCreate "github.com/stirboy/jh/pkg/cmd/jira/create"
	"github.com/stirboy/jh/pkg/factory"
)

func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jh",
		Short: "Jira Helper",
		Example: heredoc.Doc(`
			$ jh auth
			$ jh create
			$ jh create -b branch-name

			# Most common case (for more info run 'jh cr --help')
			# 1. Create jira issue
			# 2. Create new branch and checkout on it (@ is substituted with created issue key)
			$ jh cr -b @ 
		`),
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	cmd.AddCommand(auth.NewAuthCmd(f))
	cmd.AddCommand(jiraCreate.NewCreateCmd(f))

	auth.DisableAuthCheck(cmd)

	return cmd
}
