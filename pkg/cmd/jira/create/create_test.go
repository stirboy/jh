package create

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/google/shlex"
	"github.com/stirboy/jh/pkg/cmd/jira/gitclient"
	"github.com/stirboy/jh/pkg/cmd/jira/prompt"
	"github.com/stirboy/jh/pkg/cmd/jira/tests/httpmock"
	"github.com/stirboy/jh/pkg/config"
	"github.com/stirboy/jh/pkg/factory"
	"github.com/stirboy/jh/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/trivago/tgo/tcontainer"
)

func runCreateCommand(f *factory.Factory, args ...string) error {
	cmd := NewCreateCmd(f)
	cmd.SetArgs(args)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	_, err := cmd.ExecuteC()
	return err
}

func TestCreate_test_backward_compatible_interactive_flow(t *testing.T) {
	tests := []struct {
		name        string
		cfgF        func(*config.ConfigMock)
		httpStubs   func(*httpmock.Registry)
		promptsF    func(*prompt.PrompterMock)
		args        string
		inputCalls  int
		selectCalls int
		expectErr   bool
	}{
		{
			name:        "should create jira issue when interactive configuration was not provided",
			cfgF:        cfgStubs(),
			httpStubs:   httpStubs(),
			promptsF:    promptsStubs(),
			args:        "",
			inputCalls:  1,
			selectCalls: 2,
			expectErr:   false,
		},
		{
			name:        "should create jira issue with and checkout to new branch (shorthand)",
			cfgF:        cfgStubs(),
			httpStubs:   httpStubs(),
			promptsF:    promptsStubs(),
			args:        "-b some-name",
			inputCalls:  1,
			selectCalls: 2,
			expectErr:   false,
		},
		{
			name:        "should create jira issue with and checkout to new branch",
			cfgF:        cfgStubs(),
			httpStubs:   httpStubs(),
			promptsF:    promptsStubs(),
			args:        "--branch some-name",
			inputCalls:  1,
			selectCalls: 2,
			expectErr:   false,
		},
		{
			name:      "should error on flag without branch name (shorthand)",
			args:      "-b",
			expectErr: true,
		},
		{
			name:      "should error on flag without branch name",
			args:      "--branch",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		readConfigF := config.StubWriteConfig(t)
		cfg := config.NewBlankConfig()
		cfg.Set("url", "url")
		cfg.Set("username", "username")
		cfg.Set("token", "token")

		if tt.cfgF != nil {
			tt.cfgF(cfg)
		}

		t.Run(tt.name, func(t *testing.T) {
			// given

			// apply request stubs
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			if tt.httpStubs != nil {
				tt.httpStubs(reg)
			}

			// apply prompter stubs
			p := &prompt.PrompterMock{}
			if tt.promptsF != nil {
				tt.promptsF(p)
			}

			// create factory
			out := &bytes.Buffer{}
			factory := &factory.Factory{
				Config: func() (config.Config, error) {
					return cfg, nil
				},
				JiraClient: func() (*jira.Client, error) {
					// todo: provide jira client stub
					c := &http.Client{
						Transport: reg,
					}
					return jira.NewClient("https://jira-url", c)
				},
				Prompter: p,
				GitClient: func() (gitclient.GitClient, error) {
					return gitclient.NewGitClientMock(), nil
				},
				IOStream: &iostreams.IOStream{
					Out: out,
				},
			}

			argv, err := shlex.Split(tt.args)
			assert.NoError(t, err)

			// when
			err = runCreateCommand(factory, argv...)

			// then
			assert := assert.New(t)

			if tt.expectErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tt.inputCalls != 0 {
				assert.Equal(tt.inputCalls, len(p.InputCalls()))
				assert.Equal("Issue Summary", p.InputCalls()[0].S1)
			}

			if tt.selectCalls != 0 {
				assert.Equal(tt.selectCalls, len(p.SelectCalls()))
				assert.Equal("Pick a project", p.SelectCalls()[0].S)
				assert.Equal("Pick issue type", p.SelectCalls()[1].S)
			}

			assert.Equal("Seems like non-interactive mode was not configured. Running interactively...\n\ncreated issue: https://jira-url/browse/PROJ-1\n", out.String())

			outBuf := bytes.Buffer{}
			readConfigF(&outBuf)
			assert.Equal("url: url\nusername: username\ntoken: token\nconfiguration:\n    issue:\n        projectKey: PROJ\n        issueTypeName: Task\n", outBuf.String())
		})
	}
}

func TestCreate_test_interactive_flow(t *testing.T) {
	tests := []struct {
		name        string
		cfgF        func(*config.ConfigMock)
		httpStubs   func(*httpmock.Registry)
		promptsF    func(*prompt.PrompterMock)
		args        string
		inputCalls  int
		selectCalls int
		expectErr   bool
	}{
		{
			name:        "should create jira issue",
			cfgF:        cfgStubs(),
			httpStubs:   httpStubs(),
			promptsF:    promptsStubs(),
			args:        "--interactive",
			inputCalls:  1,
			selectCalls: 2,
			expectErr:   false,
		},
		{
			name:        "should create jira issue with and checkout to new branch (shorthand)",
			cfgF:        cfgStubs(),
			httpStubs:   httpStubs(),
			promptsF:    promptsStubs(),
			args:        "-i -b some-name",
			inputCalls:  1,
			selectCalls: 2,
			expectErr:   false,
		},
		{
			name:        "should create jira issue with and checkout to new branch",
			cfgF:        cfgStubs(),
			httpStubs:   httpStubs(),
			promptsF:    promptsStubs(),
			args:        "--interactive --branch some-name",
			inputCalls:  1,
			selectCalls: 2,
			expectErr:   false,
		},
		{
			name:      "should error on flag without branch name (shorthand)",
			args:      "-i -b",
			expectErr: true,
		},
		{
			name:      "should error on flag without branch name",
			args:      "--branch",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		readConfigF := config.StubWriteConfig(t)
		cfg := config.NewBlankConfig()

		if tt.cfgF != nil {
			tt.cfgF(cfg)
		}

		t.Run(tt.name, func(t *testing.T) {
			// given

			// apply request stubs
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			if tt.httpStubs != nil {
				tt.httpStubs(reg)
			}

			// apply prompter stubs
			p := &prompt.PrompterMock{}
			if tt.promptsF != nil {
				tt.promptsF(p)
			}

			// create factory
			out := &bytes.Buffer{}
			factory := &factory.Factory{
				Config: func() (config.Config, error) {
					return cfg, nil
				},
				JiraClient: func() (*jira.Client, error) {
					// todo: provide jira client stub
					c := &http.Client{
						Transport: reg,
					}
					return jira.NewClient("https://jira-url", c)
				},
				Prompter: p,
				GitClient: func() (gitclient.GitClient, error) {
					return gitclient.NewGitClientMock(), nil
				},
				IOStream: &iostreams.IOStream{
					Out: out,
				},
			}

			argv, err := shlex.Split(tt.args)
			assert.NoError(t, err)

			// when
			err = runCreateCommand(factory, argv...)

			// then
			assert := assert.New(t)

			if tt.expectErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tt.inputCalls != 0 {
				assert.Equal(tt.inputCalls, len(p.InputCalls()))
				assert.Equal("Issue Summary", p.InputCalls()[0].S1)
			}

			if tt.selectCalls != 0 {
				assert.Equal(tt.selectCalls, len(p.SelectCalls()))
				assert.Equal("Pick a project", p.SelectCalls()[0].S)
				assert.Equal("Pick issue type", p.SelectCalls()[1].S)
			}

			assert.Equal("Seems like non-interactive mode was not configured. Running interactively...\n\ncreated issue: https://jira-url/browse/PROJ-1\n", out.String())

			outBuf := bytes.Buffer{}
			readConfigF(&outBuf)
			assert.Equal("configuration:\n    issue:\n        projectKey: PROJ\n        issueTypeName: Task\n", outBuf.String())
		})
	}
}

func TestCreate_test_non_interactive_flow(t *testing.T) {
	tests := []struct {
		name        string
		cfgF        func(*config.ConfigMock)
		httpStubs   func(*httpmock.Registry)
		promptsF    func(*prompt.PrompterMock)
		args        string
		inputCalls  int
		selectCalls int
		expectErr   bool
	}{
		{
			name:        "should create jira issue",
			cfgF:        cfgStubs(),
			httpStubs:   nonInteractiveHttpStub(),
			promptsF:    promptsStubs(),
			args:        "",
			inputCalls:  1,
			selectCalls: 0,
			expectErr:   false,
		},
		{
			name:        "should create jira issue with and checkout to new branch (shorthand)",
			cfgF:        cfgStubs(),
			httpStubs:   nonInteractiveHttpStub(),
			promptsF:    promptsStubs(),
			args:        "-b some-name",
			inputCalls:  1,
			selectCalls: 0,
			expectErr:   false,
		},
		{
			name:        "should create jira issue with and checkout to new branch",
			cfgF:        cfgStubs(),
			httpStubs:   nonInteractiveHttpStub(),
			promptsF:    promptsStubs(),
			args:        "--branch some-name",
			inputCalls:  1,
			selectCalls: 0,
			expectErr:   false,
		},
		{
			name:      "should error on flag without branch name (shorthand)",
			args:      "-b",
			expectErr: true,
		},
		{
			name:      "should error on flag without branch name",
			args:      "--branch",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		//readConfigF := config.StubWriteConfig(t)
		cfg := config.NewBlankConfig()
		cfg.Set("url", "url")
		cfg.Set("username", "username")
		cfg.Set("token", "token")
		cfg.SetNested([]string{"configuration", "issue", "projectKey"}, "PROJ")
		cfg.SetNested([]string{"configuration", "issue", "issueTypeName"}, "Task")

		if tt.cfgF != nil {
			tt.cfgF(cfg)
		}

		t.Run(tt.name, func(t *testing.T) {
			// given

			// apply request stubs
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			if tt.httpStubs != nil {
				tt.httpStubs(reg)
			}

			// apply prompter stubs
			p := &prompt.PrompterMock{}
			if tt.promptsF != nil {
				tt.promptsF(p)
			}

			// create factory
			out := &bytes.Buffer{}
			factory := &factory.Factory{
				Config: func() (config.Config, error) {
					return cfg, nil
				},
				JiraClient: func() (*jira.Client, error) {
					// todo: provide jira client stub
					c := &http.Client{
						Transport: reg,
					}
					return jira.NewClient("https://jira-url", c)
				},
				Prompter: p,
				GitClient: func() (gitclient.GitClient, error) {
					return gitclient.NewGitClientMock(), nil
				},
				IOStream: &iostreams.IOStream{
					Out: out,
				},
			}

			argv, err := shlex.Split(tt.args)
			assert.NoError(t, err)

			// when
			err = runCreateCommand(factory, argv...)

			// then
			assert := assert.New(t)

			if tt.expectErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tt.inputCalls != 0 {
				assert.Equal(tt.inputCalls, len(p.InputCalls()))
				assert.Equal("Issue Summary", p.InputCalls()[0].S1)
			}

			assert.Equal("\ncreated issue: https://jira-url/browse/PROJ-1\n", out.String())
		})
	}
}

func TestCreate_normalizeBranchName(t *testing.T) {
	tests := []struct {
		name     string
		bName    string
		expected string
	}{
		{
			name:     "should return branch without changes",
			bName:    "feature/test",
			expected: "feature/test",
		},
		{
			name:     "should return branch with jira issue",
			bName:    "feature/@/test",
			expected: "feature/issue-1/test",
		},
		{
			name:     "should return branch with jira issue at the the start",
			bName:    "@/feature/test/@/@",
			expected: "issue-1/feature/test/@/@",
		},
		{
			name:     "should return jira issue as branch name",
			bName:    "@",
			expected: "issue-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// when
			result := normalizeBranchName(tt.bName, "issue-1")

			// then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func cfgStubs() func(cm *config.ConfigMock) {
	return func(cm *config.ConfigMock) {
		cm.AuthTokenFunc = func() (string, error) {
			return "token", nil
		}
		cm.GetFunc = func(s string) (string, error) {
			switch s {
			case "url":
				return "some url", nil
			case "username":
				return "some username", nil
			}

			return "", nil
		}
	}
}

func httpStubs() func(*httpmock.Registry) {
	return func(r *httpmock.Registry) {
		r.Register(
			httpmock.REST("GET", "rest/api/3/myself"),
			httpmock.JSONResponse(&jira.User{}),
		)

		values := url.Values{
			"expand": []string{"issueTypes"},
		}
		r.Register(
			httpmock.QueryMatcher("GET", "rest/api/3/project/recent", values),
			httpmock.JSONResponse(jira.ProjectList{
				{
					Key:             "PROJ",
					Name:            "Project",
					AvatarUrls:      jira.AvatarUrls{},
					ProjectCategory: jira.ProjectCategory{},
					IssueTypes: []jira.IssueType{
						{
							Name: "Task",
						},
					},
				},
			}),
		)
		m := tcontainer.NewMarshalMap()
		taskFields := tcontainer.NewMarshalMap()
		taskFields["required"] = "true"
		m.Set("Project", taskFields)

		r.Register(
			httpmock.REST("GET", "rest/api/2/issue/createmeta"),
			httpmock.JSONResponse(&jira.CreateMetaInfo{
				Projects: []*jira.MetaProject{
					{
						IssueTypes: []*jira.MetaIssueType{
							{
								Fields: m,
							},
						},
					},
				},
			}),
		)

		r.Register(
			httpmock.REST("POST", "rest/api/2/issue"),
			httpmock.JSONResponse(&jira.Issue{Key: "PROJ-1"}),
		)
	}
}

func nonInteractiveHttpStub() func(*httpmock.Registry) {
	return func(r *httpmock.Registry) {
		r.Register(
			httpmock.REST("GET", "rest/api/3/myself"),
			httpmock.JSONResponse(&jira.User{}),
		)
		m := tcontainer.NewMarshalMap()
		taskFields := tcontainer.NewMarshalMap()
		taskFields["required"] = "true"
		m.Set("Project", taskFields)

		r.Register(
			httpmock.REST("GET", "rest/api/2/issue/createmeta"),
			httpmock.JSONResponse(&jira.CreateMetaInfo{
				Projects: []*jira.MetaProject{
					{
						IssueTypes: []*jira.MetaIssueType{
							{
								Fields: m,
							},
						},
					},
				},
			}),
		)

		r.Register(
			httpmock.REST("POST", "rest/api/2/issue"),
			httpmock.JSONResponse(&jira.Issue{Key: "PROJ-1"}),
		)
	}
}

func promptsStubs() func(pm *prompt.PrompterMock) {
	return func(pm *prompt.PrompterMock) {
		pm.InputFunc = func(s1, s2 string, askOpts ...survey.AskOpt) (string, error) {
			return "This is a summary of an issue", nil
		}
		pm.SelectFunc = func(s string, strings []string) (string, error) {
			return strings[0], nil
		}
	}
}
