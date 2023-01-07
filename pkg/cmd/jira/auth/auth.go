package auth

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/cmd/jira/prompt"
	"github.com/stirboy/jh/pkg/cmd/jira/users"
	"github.com/stirboy/jh/pkg/config"
	"github.com/stirboy/jh/pkg/factory"
	"github.com/stirboy/jh/pkg/utils"
)

type AuthOptions struct {
	Config     func() (config.Config, error)
	JiraClient func() (*jira.Client, error)
	Prompter   prompt.Prompter
}

func NewAuthCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "jira authentication",

		RunE: func(cmd *cobra.Command, args []string) error {
			ops := &AuthOptions{
				Config:     f.Config,
				JiraClient: f.JiraClient,
				Prompter:   f.Prompter,
			}
			return run(ops)
		},
	}

	DisableAuthCheck(cmd)

	return cmd
}

func run(ops *AuthOptions) error {

	cfg, err := ops.Config()
	if err != nil {
		return err
	}

	t, err := cfg.AuthToken()
	if err != nil {
		return err
	}
	if t != "" {
		reauthenticate, err := ops.Prompter.Confirm("You have already been authenticated with jira. Do you want to reauthenticate?")
		if err != nil {
			return err
		}

		if !reauthenticate {
			return nil
		}
	}

	url, err := ops.Prompter.InputWithHelp("url", "",
		"Here you should type jira url. Ex. https://my-company.attlasian.net",
		survey.WithValidator(survey.Required))
	if err != nil {
		return err
	}

	username, err := ops.Prompter.InputWithHelp("username", "",
		"Here you should type jira username. Ex. my-name@gmail.com",
		survey.WithValidator(survey.Required))
	if err != nil {
		return err
	}

	token, err := ops.Prompter.InputWithHelp("token", "",
		"Here you should type jira API token. You can generate one here  https://id.atlassian.com/manage-profile/security/api-tokens",
		survey.WithValidator(survey.Required))
	if err != nil {
		return err
	}

	cfg.Set("url", url)
	cfg.Set("username", username)
	cfg.Set("token", token)
	if err := cfg.Write(); err != nil {
		return err
	}

	jiraClient, err := ops.JiraClient()
	if err != nil {
		return err
	}

	_, resp, err := users.GetCurrentUser(jiraClient)
	if err != nil {
		if resp != nil {
			return utils.ParseJiraResponse(resp)
		}
		return err
	}
	fmt.Println("Successfully authenticated.")
	return nil
}
