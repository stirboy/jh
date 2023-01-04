package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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

	curUser, _, err := jiraClient.User.GetCurrentUser(context.Background())
	if err != nil {
		return err
	}

	projects, _, err := jiraClient.Project.GetAll(context.Background(), &jira.GetQueryOptions{
		Expand: "issueTypes",
	})
	if err != nil {
		return err
	}

	summary := getSummary()
	project := getProject(projects)
	issueType := getIssueType(project)

	issue, resp, err := jiraClient.Issue.Create(context.Background(), &jira.Issue{
		Fields: &jira.IssueFields{
			Summary:  summary,
			Reporter: curUser,
			Assignee: curUser,
			Project: jira.Project{
				Key: project.Key,
			},
			Type: jira.IssueType{
				Name: issueType.Name,
			},
		},
	})
	if err != nil {
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.New("can't parse response body")
		}
		bodyString := string(bodyBytes)
		return errors.New(bodyString)
	}

	fmt.Printf("Created Issue: %s%s%s\n", jiraClient.BaseURL, "browse/", issue.Key)
	return nil
}

func getSummary() string {
	summary := ""
	prompt := &survey.Input{
		Message: "Issue Summary",
	}
	askSurvey(prompt, &summary)

	return summary
}

type Project struct {
	Key            string
	ProjectTypeKey string
	IssueTypes     []jira.IssueType
}

func getProject(projects *jira.ProjectList) *Project {
	mapOfProjects := make(map[string]*Project)
	for i := 0; i < len(*projects); i++ {
		mapOfProjects[(*projects)[i].Key] = &Project{
			Key:            (*projects)[i].Key,
			ProjectTypeKey: (*projects)[i].ProjectTypeKey,
			IssueTypes:     (*projects)[i].IssueTypes,
		}
	}

	projectKeys := make([]string, len(mapOfProjects))
	idx := 0
	for k := range mapOfProjects {
		projectKeys[idx] = k
		idx++
	}

	prompt := &survey.Select{
		Message:  "Pick a project",
		Options:  projectKeys,
		PageSize: 10,
	}

	selectedProject := ""
	survey.AskOne(prompt, &selectedProject)

	return mapOfProjects[selectedProject]
}

func getIssueType(project *Project) *jira.IssueType {
	mapOfIssueTypes := make(map[string]*jira.IssueType)
	for i := 0; i < len(project.IssueTypes); i++ {
		mapOfIssueTypes[project.IssueTypes[i].Name] = &project.IssueTypes[i]
	}

	issueTypeKeys := make([]string, len(mapOfIssueTypes))
	idx := 0
	for k := range mapOfIssueTypes {
		issueTypeKeys[idx] = k
		idx++
	}

	prompt := &survey.Select{
		Message:  "Pick a issue type",
		Options:  issueTypeKeys,
		PageSize: 10,
	}

	selectedIssueType := ""
	survey.AskOne(prompt, &selectedIssueType)

	return mapOfIssueTypes[selectedIssueType]
}

func askSurvey(p survey.Prompt, r interface{}) {
	survey.AskOne(p, r, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = ">>"
		icons.Question.Format = "green+hb"
	}), survey.WithValidator(func(ans interface{}) error {
		return Validator(ans.(string))
	}),
	)
}

func Validator(value string) error {
	if len(strings.TrimSpace(value)) < 1 {
		return errors.New("a value is required")
	}
	return nil
}
