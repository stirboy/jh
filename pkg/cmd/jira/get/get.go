package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/factory"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

type GetOptions struct {
	JiraIssueKey string
	JiraClient   func() (*jira.Client, error)
}

func NewGetCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <jira-key>",
		Short: "get jira issue",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			ops := &GetOptions{
				JiraIssueKey: args[0],
				JiraClient:   f.JiraClient,
			}
			return run(ops)
		},
	}

	return cmd
}

func run(ops *GetOptions) error {
	jiraClient, err := ops.JiraClient()
	if err != nil {
		return err
	}

	issue, _, err := jiraClient.Issue.Get(context.Background(), ops.JiraIssueKey, &jira.GetQueryOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Issue key: %v\n", issue.Fields.Reporter)
	return nil
}
