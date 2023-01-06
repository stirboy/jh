package create

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

type CurrentUserResult struct {
	user *jira.User
	err  error
}

func getCurrentUserResultAsync(jiraClient *jira.Client, ch chan<- *CurrentUserResult) {
	go func() {
		result := getCurrentUser(jiraClient)
		ch <- result
		close(ch)
	}()
}

func getCurrentUser(jiraClient *jira.Client) *CurrentUserResult {
	u, _, err := jiraClient.User.GetCurrentUser(context.Background())
	if err != nil {
		return &CurrentUserResult{
			user: nil,
			err:  err,
		}
	}

	return &CurrentUserResult{
		user: u,
		err:  err,
	}
}
