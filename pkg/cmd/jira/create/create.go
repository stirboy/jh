package create

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/cmd/jira/prompt"
	"github.com/stirboy/jh/pkg/cmd/jira/users"
	"github.com/stirboy/jh/pkg/factory"
	"github.com/stirboy/jh/pkg/utils"
)

type CreateOptions struct {
	JiraClient      func() (*jira.Client, error)
	Prompter        prompt.Prompter
	CreateGitBranch string
}

func NewGetCmd(f *factory.Factory) *cobra.Command {
	ops := &CreateOptions{
		JiraClient: f.JiraClient,
		Prompter:   f.Prompter,
	}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "create jira issue",
		Args:    cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ops)
		},
	}

	cmd.Flags().StringVarP(&ops.CreateGitBranch, "new-branch", "b", "", "Create new branch with newly created jira issue key")

	return cmd
}

func run(ops *CreateOptions) error {
	jiraClient, err := ops.JiraClient()
	if err != nil {
		return err
	}

	// get current user without blocking the flow
	curUserChan := make(chan *users.CurrentUserResult)
	users.GetCurrentUserResultAsync(jiraClient, curUserChan)

	// get recent project without blocking the flow
	recentProjectsResultChan := make(chan *ProjectResult)
	getRecentProjectsResultAsync(jiraClient, recentProjectsResultChan)

	summary, err := inputSummary(ops.Prompter)
	if err != nil {
		return err
	}

	// waiting for recent project to load
	recentProjectResult := <-recentProjectsResultChan
	if err = recentProjectResult.err; err != nil {
		return err
	}

	project, err := selectProject(ops.Prompter, recentProjectResult.projectKeyMap)
	if err != nil {
		return err
	}

	// get required fields for jira project without blocking the flow
	requiredFieldsChan := make(chan *RequiredFieldsResult)
	getRequiredFieldsResultAsync(jiraClient, project, requiredFieldsChan)

	issueType, err := selectIssueType(ops.Prompter, project)
	if err != nil {
		return err
	}

	// waiting for required fields to load
	requiredFieldsResult := <-requiredFieldsChan
	if err = requiredFieldsResult.err; err != nil {
		return err
	}

	// waiting fir current user to load
	currentUserResult := <-curUserChan
	if err = currentUserResult.Err; err != nil {
		return err
	}

	var reporter *jira.User
	reporterRequired := requiredFieldsResult.fields["reporter"]
	if reporterRequired {
		reporter = currentUserResult.User
	}

	issue, resp, err := jiraClient.Issue.Create(context.Background(), &jira.Issue{
		Fields: &jira.IssueFields{
			Summary:  summary,
			Reporter: reporter,
			Assignee: currentUserResult.User,
			Project: jira.Project{
				Key: project.Key,
			},
			Type: jira.IssueType{
				Name: issueType.Name,
			},
		},
	})
	if err != nil {
		return utils.ParseJiraResponse(resp)
	}

	fmt.Printf("\ncreated issue: %s%s%s\n", jiraClient.BaseURL, "browse/", issue.Key)

	if err = CreateGitBranch(issue.Key, ops); err != nil {
		return err
	}
	return nil
}
