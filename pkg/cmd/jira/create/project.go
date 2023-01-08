package create

import (
	"context"
	"net/http"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

type Project struct {
	Key        string
	IssueTypes []jira.IssueType
}

type ProjectResult struct {
	projectKeyMap map[string]*Project
	err           error
}

func getRecentProjectsResultAsync(jiraClient *jira.Client, ch chan<- *ProjectResult) {
	go func() {
		result := getProjectResult(jiraClient)
		ch <- result
		close(ch)
	}()
}

func getProjectResult(jiraClient *jira.Client) *ProjectResult {
	req, err := jiraClient.NewRequest(context.Background(),
		http.MethodGet, "rest/api/3/project/recent?expand=issueTypes", nil)
	if err != nil {
		return &ProjectResult{
			projectKeyMap: nil,
			err:           err,
		}
	}

	projects := new(jira.ProjectList)
	_, err = jiraClient.Do(req, projects)
	if err != nil {
		return &ProjectResult{
			projectKeyMap: nil,
			err:           err,
		}
	}

	mapOfProjects := make(map[string]*Project)
	for i := 0; i < len(*projects); i++ {
		mapOfProjects[(*projects)[i].Key] = &Project{
			Key:        (*projects)[i].Key,
			IssueTypes: (*projects)[i].IssueTypes,
		}
	}

	return &ProjectResult{
		projectKeyMap: mapOfProjects,
		err:           nil,
	}
}
