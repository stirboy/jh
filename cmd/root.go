package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/cmd/jira/auth"
	jiraCreate "github.com/stirboy/jh/pkg/cmd/jira/create"
	jiraGet "github.com/stirboy/jh/pkg/cmd/jira/get"
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
		`),
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	cmd.AddCommand(auth.NewAuthCmd(f))
	cmd.AddCommand(jiraGet.NewGetCmd(f))
	cmd.AddCommand(jiraCreate.NewCreateCmd(f))

	auth.DisableAuthCheck(cmd)

	return cmd
}
