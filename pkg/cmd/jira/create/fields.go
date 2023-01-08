package create

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

type RequiredFieldsResult struct {
	fields map[string]bool
	err    error
}

func getRequiredFieldsResultAsync(c *jira.Client, p *Project, ch chan<- *RequiredFieldsResult) {
	go func() {
		result := getRequiredFieldsResult(c, p)
		ch <- result
		close(ch)
	}()
}

func getRequiredFieldsResult(jiraClient *jira.Client, project *Project) *RequiredFieldsResult {
	fields, _, err := jiraClient.Issue.GetCreateMeta(context.Background(), &jira.GetQueryOptions{
		Expand:      "projects.issuetypes.fields",
		ProjectKeys: project.Key,
	})
	if err != nil {
		return &RequiredFieldsResult{
			fields: nil,
			err:    err,
		}
	}

	m := fields.Projects[0].IssueTypes[0].Fields
	requiredFields := make(map[string]bool)
	for k := range m {
		isRequired, exists := m.Value(k + "/required")
		if exists && isRequired == true {
			requiredFields[k] = true
		}
	}

	return &RequiredFieldsResult{
		fields: requiredFields,
		err:    nil,
	}
}
