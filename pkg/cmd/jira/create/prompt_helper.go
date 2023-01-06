package create

import (
	"github.com/AlecAivazis/survey/v2"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/stirboy/jh/pkg/cmd/jira/prompt"
	"github.com/stirboy/jh/pkg/utils"
)

func inputSummary(prompter prompt.Prompter) (string, error) {
	summary, err := prompter.Input("Issue Summary", "Minor fixes", survey.WithValidator(survey.Required))
	if err != nil {
		return "", err
	}

	return summary, nil
}

func selectProject(prompter prompt.Prompter, projectKeyMap map[string]*Project) (*Project, error) {
	projectKeys, err := utils.MapStringKeys(projectKeyMap)
	if err != nil {
		return nil, err
	}

	p, err := prompter.Select("Pick a project", projectKeys)
	if err != nil {
		return nil, err
	}

	return projectKeyMap[p], nil
}

func selectIssueType(prompter prompt.Prompter, project *Project) (*jira.IssueType, error) {
	mapOfIssueTypes := make(map[string]*jira.IssueType)
	for i := 0; i < len(project.IssueTypes); i++ {
		mapOfIssueTypes[project.IssueTypes[i].Name] = &project.IssueTypes[i]
	}

	issueTypeKeys, err := utils.MapStringKeys(mapOfIssueTypes)
	if err != nil {
		return nil, err
	}

	t, err := prompter.Select("Pick issue type", issueTypeKeys)
	if err != nil {
		return nil, err
	}

	return mapOfIssueTypes[t], nil
}
