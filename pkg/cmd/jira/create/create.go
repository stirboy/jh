package create

import (
	"context"
	"errors"
	"fmt"
	"io"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/factory"
)

type CreateOptions struct {
	JiraClient       func() (*jira.Client, error)
	IssueSummary     string
	IssueDescription string
	JiraIssueType    string
}

func NewGetCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "create jira issue",
		Args:    cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			ops := &CreateOptions{
				JiraClient: f.JiraClient,
			}
			return run(ops)
		},
	}

	return cmd
}

func run(ops *CreateOptions) error {
	jiraClient, err := ops.JiraClient()
	if err != nil {
		return err
	}

	// get current user without blocking the flow
	curUserChan := make(chan *CurrentUserResult)
	getCurrentUserResultAsync(jiraClient, curUserChan)

	// get recent project without blocking the flow
	recentProjectsResultChan := make(chan *ProjectResult)
	getRecentProjectsResultAsync(jiraClient, recentProjectsResultChan)

	summary, err := selectSummary()
	if err != nil {
		return err
	}

	// waiting for recent project to load
	recentProjectResult := <-recentProjectsResultChan
	if recentProjectResult.err != nil {
		return err
	}

	project, err := selectProject(recentProjectResult.projectKeyMap)
	if err != nil {
		return err
	}

	// get required fields for jira project without blocking the flow
	requiredFieldsChan := make(chan *RequiredFieldsResult)
	getRequiredFieldsResultAsync(jiraClient, project, requiredFieldsChan)

	issueType, err := selectIssueType(project)
	if err != nil {
		return err
	}

	// waiting for required fields to load
	requiredFieldsResult := <-requiredFieldsChan
	if requiredFieldsResult.err != nil {
		return requiredFieldsResult.err
	}

	// waiting fir current user to load
	currentUserResult := <-curUserChan
	if currentUserResult.err != nil {
		return currentUserResult.err
	}

	var reporter *jira.User
	reporterRequired := requiredFieldsResult.fields["reporter"]
	if reporterRequired {
		reporter = currentUserResult.user
	}

	issue, resp, err := jiraClient.Issue.Create(context.Background(), &jira.Issue{
		Fields: &jira.IssueFields{
			Summary:  summary,
			Reporter: reporter,
			Assignee: currentUserResult.user,
			Project: jira.Project{
				Key: project.Key,
			},
			Type: jira.IssueType{
				Name: issueType.Name,
			},
		},
	})
	if err != nil {
		return parseCreateResponse(resp)
	}

	fmt.Printf("Created Issue: %s%s%s\n", jiraClient.BaseURL, "browse/", issue.Key)
	return nil
}

func parseCreateResponse(resp *jira.Response) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("can't parse response body")
	}
	bodyString := string(bodyBytes)
	return errors.New(bodyString)
}
