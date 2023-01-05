package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
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

type ProjectResult struct {
	projectMap map[string]*Project
	err        error
}

func obtainProject(jiraClient *jira.Client) *ProjectResult {
	req, err := jiraClient.NewRequest(context.Background(),
		http.MethodGet, "/rest/api/3/project/recent?expand=issueTypes",
		&jira.GetQueryOptions{
			Expand: "issueTypes",
		})
	if err != nil {
		return &ProjectResult{
			projectMap: nil,
			err:        err,
		}
	}

	projects := new(jira.ProjectList)
	_, err = jiraClient.Do(req, projects)
	if err != nil {
		return &ProjectResult{
			projectMap: nil,
			err:        err,
		}
	}

	mapOfProjects := make(map[string]*Project)
	for i := 0; i < len(*projects); i++ {
		mapOfProjects[(*projects)[i].Key] = &Project{
			Key:            (*projects)[i].Key,
			ProjectTypeKey: (*projects)[i].ProjectTypeKey,
			IssueTypes:     (*projects)[i].IssueTypes,
		}
	}

	return &ProjectResult{
		projectMap: mapOfProjects,
		err:        nil,
	}
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

	projectResultChan := make(chan *ProjectResult)
	go func() {
		p := obtainProject(jiraClient)
		projectResultChan <- p
		close(projectResultChan)
	}()

	summary, err := selectSummary()
	if err != nil {
		return err
	}

	projectResult := <-projectResultChan
	if projectResult.err != nil {
		return err
	}

	project, err := selectProject(projectResult.projectMap)
	if err != nil {
		return err
	}

	issueType, err := selectIssueType(project)
	if err != nil {
		return err
	}

	fields, _, err := jiraClient.Issue.GetCreateMeta(context.Background(), &jira.GetQueryOptions{
		Expand:      "projects.issuetypes.fields",
		ProjectKeys: project.Key,
	})
	if err != nil {
		return err
	}

	m := fields.Projects[0].IssueTypes[0].Fields
	requiredFields := make(map[string]bool)
	for k := range m {
		r, exists := m.Value(k + "/required")
		if exists && r == true {
			requiredFields[k] = true
		}
	}

	fmt.Printf("required fields %v\n", requiredFields)

	var reporter *jira.User
	reporterRequired := requiredFields["reporter"]
	if reporterRequired {
		reporter = curUser
	}

	issue, resp, err := jiraClient.Issue.Create(context.Background(), &jira.Issue{
		Fields: &jira.IssueFields{
			Summary:  summary,
			Reporter: reporter,
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

func selectSummary() (string, error) {
	summary := ""
	prompt := &survey.Input{
		Message: "Issue Summary",
	}
	if err := askSurvey(prompt, &summary); err != nil {
		return "", err
	}

	return summary, nil
}

type Project struct {
	Key            string
	ProjectTypeKey string
	IssueTypes     []jira.IssueType
}

func selectProject(mapOfProjects map[string]*Project) (*Project, error) {
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
	err := survey.AskOne(prompt, &selectedProject)
	if err != nil {
		return nil, err
	}

	return mapOfProjects[selectedProject], nil
}

func selectIssueType(project *Project) (*jira.IssueType, error) {
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
	err := survey.AskOne(prompt, &selectedIssueType)
	if err != nil {
		return nil, err
	}

	return mapOfIssueTypes[selectedIssueType], nil
}

func askSurvey(p survey.Prompt, r interface{}) error {
	return survey.AskOne(p, r, survey.WithIcons(func(icons *survey.IconSet) {
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
