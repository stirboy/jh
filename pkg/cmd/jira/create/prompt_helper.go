package create

import (
	"errors"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/stirboy/jh/pkg/utils"
)

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

func selectProject(projectKeyMap map[string]*Project) (*Project, error) {
	projectKeys, err := utils.MapStringKeys(projectKeyMap)
	if err != nil {
		return nil, err
	}

	prompt := &survey.Select{
		Message:  "Pick a project",
		Options:  projectKeys,
		PageSize: 10,
	}

	selectedProject := ""
	err = survey.AskOne(prompt, &selectedProject)
	if err != nil {
		return nil, err
	}

	return projectKeyMap[selectedProject], nil
}

func selectIssueType(project *Project) (*jira.IssueType, error) {
	mapOfIssueTypes := make(map[string]*jira.IssueType)
	for i := 0; i < len(project.IssueTypes); i++ {
		mapOfIssueTypes[project.IssueTypes[i].Name] = &project.IssueTypes[i]
	}

	issueTypeKeys, err := utils.MapStringKeys(mapOfIssueTypes)
	if err != nil {
		return nil, err
	}

	prompt := &survey.Select{
		Message:  "Pick a issue type",
		Options:  issueTypeKeys,
		PageSize: 10,
	}

	selectedIssueType := ""
	err = survey.AskOne(prompt, &selectedIssueType)
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
