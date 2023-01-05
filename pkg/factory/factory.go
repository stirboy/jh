package factory

import (
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/stirboy/jh/pkg/config"
)

type Factory struct {
	Config     func() (config.Config, error)
	JiraClient func() (*jira.Client, error)
}

func NewFactory() *Factory {
	f := &Factory{
		Config: configFunc(),
	}

	f.JiraClient = jiraClientFunc(f) // depends on Config

	return f
}

func configFunc() func() (config.Config, error) {
	var cachedConfig config.Config
	var configError error
	return func() (config.Config, error) {
		if cachedConfig != nil || configError != nil {
			return cachedConfig, configError
		}
		cachedConfig, configError := config.NewConfig()
		return cachedConfig, configError
	}
}

func jiraClientFunc(f *Factory) func() (*jira.Client, error) {
	return func() (*jira.Client, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}

		url, err := cfg.Get("url")
		if err != nil {
			return nil, err
		}

		username, err := cfg.Get("username")
		if err != nil {
			return nil, err
		}

		token, err := cfg.Get("token")
		if err != nil {
			return nil, err
		}

		tp := jira.BasicAuthTransport{
			Username: username,
			APIToken: token,
		}

		jiraClient, err := jira.NewClient(url, tp.Client())
		if err != nil {
			return nil, err
		}

		return jiraClient, nil
	}
}
