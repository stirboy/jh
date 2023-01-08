package create

import (
	"sort"
	"strings"

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
	projectKeys, err := utils.MapKeys(projectKeyMap)
	if err != nil {
		return nil, err
	}
	sort.Strings(projectKeys)

	p, err := prompter.Select("Pick a project", projectKeys)
	if err != nil {
		return nil, err
	}

	return projectKeyMap[p], nil
}

func selectIssueType(prompter prompt.Prompter, project *Project) (*jira.IssueType, error) {
	mapOfIssueTypes := make(map[string]*jira.IssueType)
	for i := 0; i < len(project.IssueTypes); i++ {
		if strings.ToLower(project.IssueTypes[i].Name) == "subtask" {
			// let's skip creating subtask as it requires parent id
			// we can always assign it to task in jira website after creation
			continue
		}
		mapOfIssueTypes[project.IssueTypes[i].Name] = &project.IssueTypes[i]
	}

	issueTypeKeys, err := utils.MapKeys(mapOfIssueTypes)
	if err != nil {
		return nil, err
	}
	sort.Strings(issueTypeKeys)

	t, err := prompter.Select("Pick issue type", issueTypeKeys)
	if err != nil {
		return nil, err
	}

	return mapOfIssueTypes[t], nil
}
