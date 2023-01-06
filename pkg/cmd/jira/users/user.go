package users

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

type CurrentUserResult struct {
	User *jira.User
	Err  error
}

func GetCurrentUserResultAsync(jiraClient *jira.Client, ch chan<- *CurrentUserResult) {
	go func() {
		var result *CurrentUserResult
		u, _, err := GetCurrentUser(jiraClient)
		if err != nil {
			result = &CurrentUserResult{
				User: nil,
				Err:  err,
			}
		}

		result = &CurrentUserResult{
			User: u,
			Err:  err,
		}
		ch <- result
		close(ch)
	}()
}

func GetCurrentUser(jiraClient *jira.Client) (*jira.User, *jira.Response, error) {
	u, resp, err := jiraClient.User.GetCurrentUser(context.Background())
	return u, resp, err
}
