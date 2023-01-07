package auth

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/stirboy/jh/pkg/cmd/jira/prompt"
	"github.com/stirboy/jh/pkg/cmd/jira/tests/httpmock"
	"github.com/stirboy/jh/pkg/config"
	"github.com/stirboy/jh/pkg/factory"
	"github.com/stretchr/testify/assert"
)

func runAuthCommand(f *factory.Factory) error {
	cmd := NewAuthCmd(f)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	_, err := cmd.ExecuteC()
	return err
}

func TestAuth_should_return_not_found_error(t *testing.T) {
	c := heredoc.Doc(
		`
		username: ""

		`)

	cfg := config.NewFromString(c)
	factory := &factory.Factory{
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}
	err := runAuthCommand(factory)
	assert.EqualError(t, err, "not found")
}

func TestAuth_should_skip_reauthentication(t *testing.T) {

	p := &prompt.PrompterMock{
		ConfirmFunc: func(s string) (bool, error) {
			return false, nil
		},
	}
	factory := &factory.Factory{
		Config: func() (config.Config, error) {
			return &config.ConfigMock{
				AuthTokenFunc: func() (string, error) {
					return "some value", nil
				},
			}, nil
		},
		Prompter: p,
	}

	runAuthCommand(factory)

	assert.Equal(t, 1, len(p.ConfirmCalls()))
	assert.Equal(t,
		"You have already been authenticated with jira. Do you want to reauthenticate?",
		p.ConfirmCalls()[0].S)
}

func TestAuth_should_test(t *testing.T) {
	tests := []struct {
		name         string
		cfgF         func(*config.ConfigMock)
		httpStubs    func(*httpmock.Registry)
		promptsF     func(*prompt.PrompterMock)
		confirmCalls int
		inputCalls   int
	}{
		{
			name: "should authenticate user for the first time",
			cfgF: func(cm *config.ConfigMock) {
				cm.AuthTokenFunc = func() (string, error) {
					return "", nil
				}
			},
			httpStubs: func(r *httpmock.Registry) {
				r.Register(
					httpmock.REST("GET", "rest/api/3/myself"), httpmock.JSONResponse(&jira.User{}),
				)
			},
			promptsF: func(pm *prompt.PrompterMock) {
				pm.InputWithHelpFunc = func(s1, s2, s3 string, askOpts ...survey.AskOpt) (string, error) {
					switch s1 {
					case "url":
						return "url value", nil
					case "username":
						return "username value", nil
					case "token":
						return "token value", nil
					}

					return "", nil
				}
			},
			confirmCalls: 0,
			inputCalls:   3,
		},
		{
			name: "should reauthenticate user",
			cfgF: func(cm *config.ConfigMock) {
				cm.AuthTokenFunc = func() (string, error) {
					return "some value", nil
				}
			},
			httpStubs: func(r *httpmock.Registry) {
				r.Register(
					httpmock.REST("GET", "rest/api/3/myself"), httpmock.JSONResponse(&jira.User{}),
				)
			},
			promptsF: func(pm *prompt.PrompterMock) {
				pm.ConfirmFunc = func(s string) (bool, error) {
					return true, nil
				}
				pm.InputWithHelpFunc = func(s1, s2, s3 string, askOpts ...survey.AskOpt) (string, error) {
					switch s1 {
					case "url":
						return "url value", nil
					case "username":
						return "username value", nil
					case "token":
						return "token value", nil
					}

					return "", nil
				}
			},
			confirmCalls: 1,
			inputCalls:   3,
		},
	}

	for _, tt := range tests {
		readConfigF := config.StubWriteConfig(t)
		cfg := config.NewBlankConfig()

		if tt.cfgF != nil {
			tt.cfgF(cfg)
		}

		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			if tt.httpStubs != nil {
				tt.httpStubs(reg)
			}

			p := &prompt.PrompterMock{}
			if tt.promptsF != nil {
				tt.promptsF(p)
			}

			factory := &factory.Factory{
				Config: func() (config.Config, error) {
					return cfg, nil
				},
				JiraClient: func() (*jira.Client, error) {
					// todo: provide jira client stub
					c := &http.Client{
						Transport: reg,
					}
					return jira.NewClient("", c)
				},
				Prompter: p,
			}

			err := runAuthCommand(factory)

			outBuf := bytes.Buffer{}
			readConfigF(&outBuf)

			assert.NoError(t, err)
			assert.Equal(t, "url: url value\nusername: username value\ntoken: token value\n", outBuf.String())

			if tt.confirmCalls != 0 {
				assert.Equal(t, tt.confirmCalls, len(p.ConfirmCalls()))
				assert.Equal(t,
					"You have already been authenticated with jira. Do you want to reauthenticate?",
					p.ConfirmCalls()[0].S)
			}

			if tt.inputCalls != 0 {
				assert.Equal(t, tt.inputCalls, len(p.InputWithHelpCalls()))

				url := p.InputWithHelpCalls()[0]
				assert.Equal(t, "url", url.S1)
				assert.Equal(t, "", url.S2)
				assert.Equal(t, "Here you should type jira url. Ex. https://my-company.attlasian.net", url.S3)

				username := p.InputWithHelpCalls()[1]
				assert.Equal(t, "username", username.S1)
				assert.Equal(t, "", username.S2)
				assert.Equal(t, "Here you should type jira username. Ex. my-name@gmail.com", username.S3)

				token := p.InputWithHelpCalls()[2]
				assert.Equal(t, "token", token.S1)
				assert.Equal(t, "", token.S2)
				assert.Equal(t, "Here you should type jira API token. You can generate one here  https://id.atlassian.com/manage-profile/security/api-tokens", token.S3)
			}
		})
	}
}
