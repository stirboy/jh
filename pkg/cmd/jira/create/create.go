package create

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/MakeNowJust/heredoc"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/cmd/jira/gitclient"
	"github.com/stirboy/jh/pkg/cmd/jira/prompt"
	"github.com/stirboy/jh/pkg/cmd/jira/users"
	"github.com/stirboy/jh/pkg/config"
	"github.com/stirboy/jh/pkg/factory"
	"github.com/stirboy/jh/pkg/utils"
)

type CreateOptions struct {
	Config          func() (config.Config, error)
	JiraClient      func() (*jira.Client, error)
	Prompter        prompt.Prompter
	GitClient       func() (gitclient.GitClient, error)
	CreateGitBranch string
	IsInteractive   bool
	Out             io.Writer
}

func NewCreateCmd(f *factory.Factory) *cobra.Command {
	ops := &CreateOptions{
		Config:     f.Config,
		JiraClient: f.JiraClient,
		Prompter:   f.Prompter,
		GitClient:  f.GitClient,
		Out:        f.IOStream.Out,
	}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create jira issue",
		Example: heredoc.Doc(`
			# create jira issue
			$ jh create
			$ jh cr

			# create jira issue and checkout to a new branch
			$ jh create -b branch-name

			# create jira issue and checkout to a new branch which contains jira issue key
			# @ sign is replaced with actual jira issue key
			# Ex. feature/@/test --> feature/issue-1/test
			$ jh create -b @/branch-name
		`),

		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ops)
		},
	}

	cmd.Flags().StringVarP(&ops.CreateGitBranch, "branch", "b", "", "Create new branch with newly created jira issue key")
	cmd.Flags().BoolVarP(&ops.IsInteractive, "interactive", "i", false, "Provide jira details interactively")

	return cmd
}

func run(ops *CreateOptions) error {
	jiraClient, err := ops.JiraClient()
	if err != nil {
		return err
	}
	// create jira issue
	issue, err := createJiraIssue(ops)
	if err != nil {
		return err
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

func createJiraIssue(ops *CreateOptions) (*jira.Issue, error) {
	jiraClient, err := ops.JiraClient()
	if err != nil {
		return nil, err
	}

	cfg, err := ops.Config()
	if err != nil {
		return nil, err
	}

	isConfigPopulated := true
	projectKeyValue, err := cfg.GetNested([]string{"configuration", "issue", "projectKey"})
	if err != nil || projectKeyValue == "" {
		isConfigPopulated = false
	}

	if ops.IsInteractive || !isConfigPopulated {
		if projectKeyValue == "" {
			fmt.Fprintln(ops.Out, "Seems like non-interactive mode was not configured. Running interactively...")
		}
		issue, err := runInteractive(jiraClient, cfg, ops.Prompter, ops.Out)
		if err != nil {
			return nil, err
		}

		return issue, nil
	} else {
		issue, err := runNonInteractive(jiraClient, cfg, ops.Prompter)
		if err != nil {
			return nil, err
		}

		return issue, nil
	}
}

func runNonInteractive(jiraClient *jira.Client, cfg config.Config, prompter prompt.Prompter) (*jira.Issue, error) {
	// get current user without blocking the flow
	curUserChan := make(chan *users.CurrentUserResult)
	users.GetCurrentUserResultAsync(jiraClient, curUserChan)

	projectKey, err := cfg.GetNested([]string{"configuration", "issue", "projectKey"})
	if err != nil {
		return nil, err
	}

	issueTypeName, err := cfg.GetNested([]string{"configuration", "issue", "issueTypeName"})
	if err != nil {
		return nil, err
	}

	// get required fields for jira project without blocking the flow
	requiredFieldsChan := make(chan *RequiredFieldsResult)
	getRequiredFieldsResultAsync(jiraClient, &Project{Key: projectKey}, requiredFieldsChan)

	summary, err := inputSummary(prompter)
	if err != nil {
		return nil, err
	}

	// waiting fir current user to load
	currentUserResult := <-curUserChan
	if err = currentUserResult.Err; err != nil {
		return nil, err
	}

	// waiting for required fields to load
	requiredFieldsResult := <-requiredFieldsChan
	if err = requiredFieldsResult.err; err != nil {
		return nil, err
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
				Key: projectKey,
			},
			Type: jira.IssueType{
				Name: issueTypeName,
			},
		},
	})
	if err != nil {
		return nil, utils.ParseJiraResponse(resp)
	}

	return issue, nil
}

func runInteractive(jiraClient *jira.Client, cfg config.Config, prompter prompt.Prompter, out io.Writer) (*jira.Issue, error) {
	// get current user without blocking the flow
	curUserChan := make(chan *users.CurrentUserResult)
	users.GetCurrentUserResultAsync(jiraClient, curUserChan)

	// get recent project without blocking the flow
	recentProjectsResultChan := make(chan *ProjectResult)
	getRecentProjectsResultAsync(jiraClient, recentProjectsResultChan)

	summary, err := inputSummary(prompter)
	if err != nil {
		return nil, err
	}

	// waiting for recent project to load
	recentProjectResult := <-recentProjectsResultChan
	if err = recentProjectResult.err; err != nil {
		return nil, err
	}

	project, err := selectProject(prompter, recentProjectResult.projectKeyMap)
	if err != nil {
		return nil, err
	}

	// get required fields for jira project without blocking the flow
	requiredFieldsChan := make(chan *RequiredFieldsResult)
	getRequiredFieldsResultAsync(jiraClient, project, requiredFieldsChan)

	issueType, err := selectIssueType(prompter, project)
	if err != nil {
		return nil, err
	}

	// waiting for required fields to load
	requiredFieldsResult := <-requiredFieldsChan
	if err = requiredFieldsResult.err; err != nil {
		return nil, err
	}

	// waiting fir current user to load
	currentUserResult := <-curUserChan
	if err = currentUserResult.Err; err != nil {
		return nil, err
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
		return nil, utils.ParseJiraResponse(resp)
	}

	cfg.SetNested([]string{"configuration", "issue", "projectKey"}, project.Key)
	cfg.SetNested([]string{"configuration", "issue", "issueTypeName"}, issueType.Name)
	if err := cfg.Write(); err != nil {
		fmt.Fprintf(out, "Unable to populate configuration for interactive setup - %s\n", err.Error())
	}

	return issue, nil
}

func normalizeBranchName(branchName, issueKey string) string {
	return strings.Replace(branchName, "@", strings.ToLower(issueKey), 1)
}
