package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/config"
	"github.com/stirboy/jh/pkg/factory"
)

type AuthOptions struct {
	Config func() (config.Config, error)
}

func NewAuthCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "jira authentication",

		RunE: func(cmd *cobra.Command, args []string) error {
			ops := &AuthOptions{
				Config: f.Config,
			}
			return run(ops)
		},
	}
}

func run(ops *AuthOptions) error {

	cfg, err := ops.Config()
	if err != nil {
		return err
	}

	t, err := cfg.Get("token")
	if err != nil {
		return err
	}
	if t != "" {
		alreadyAuthenticated := ""
		p := &survey.Select{
			Message: "You have already been authenticated with jira. Do you want to reauthenticate?",
			Options: []string{"yes", "no"},
		}
		survey.AskOne(p, &alreadyAuthenticated)

		if alreadyAuthenticated == "no" {
			return nil
		}
	}

	url := ""
	prompt := &survey.Input{
		Message: "url",
		Help:    "Here you should type jira url. Ex. my-company.attlasian.net",
	}
	askSurvey(prompt, &url)

	username := ""
	prompt = &survey.Input{
		Message: "username",
		Help:    "Here you should type jira username. Ex. https://my-name@gmail.com",
	}
	askSurvey(prompt, &username)

	token := ""
	prompt = &survey.Input{
		Message: "token",
		Help:    "Here you should type jira API token. You can generate one here  https://id.atlassian.com/manage-profile/security/api-tokens",
	}
	askSurvey(prompt, &token)

	cfg.Set("url", url)
	cfg.Set("username", username)
	cfg.Set("token", token)
	if err := cfg.Write(); err != nil {
		return err
	}

	fmt.Print("Successfully authenticated\n")
	return nil
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
