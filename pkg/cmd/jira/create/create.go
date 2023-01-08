package create

import (
	"context"
	"fmt"
	"io"
	"strings"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/cmd/jira/gitclient"
	"github.com/stirboy/jh/pkg/cmd/jira/prompt"
	"github.com/stirboy/jh/pkg/cmd/jira/users"
	"github.com/stirboy/jh/pkg/factory"
	"github.com/stirboy/jh/pkg/utils"
)

type CreateOptions struct {
	JiraClient      func() (*jira.Client, error)
	Prompter        prompt.Prompter
	GitClient       func() (gitclient.GitClient, error)
	CreateGitBranch string
	Out             io.Writer
}

func NewCreateCmd(f *factory.Factory) *cobra.Command {
	ops := &CreateOptions{
		JiraClient: f.JiraClient,
		Prompter:   f.Prompter,
		GitClient:  f.GitClient,
		Out:        f.IOStream.Out,
	}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "create jira issue",

		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ops)
		},
	}

	cmd.Flags().StringVarP(&ops.CreateGitBranch, "branch", "b", "", "Create new branch with newly created jira issue key")

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

	fmt.Fprintf(ops.Out, "\ncreated issue: %s%s%s\n", jiraClient.BaseURL, "browse/", issue.Key)

	// create and checkout to new branch
	if ops.CreateGitBranch != "" {
		branchName := strings.Replace(ops.CreateGitBranch, "@", strings.ToLower(issue.Key), 1)

		gitClient, err := ops.GitClient()
		if err != nil {
			return err
		}

		if err = gitClient.CreateBranchWithCheckout(branchName); err != nil {
			return err
		}
	}

	return nil
}

func normalizeBranchName(branchName, issueKey string) string {
	return strings.Replace(branchName, "@", strings.ToLower(issueKey), 1)
}
