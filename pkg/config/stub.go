package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc"
)

func NewBlankConfig() *ConfigMock {
	str := heredoc.Doc(`
# What jira url to use. Ex. my-company.attlasian.net
url:
# What username to use for auth. Ex. my-email@gmail.com
username:
# Jira API Token
token:
	`)
	return NewFromString(str)
}

func NewFromString(cfgString string) *ConfigMock {
	c := ReadFromString(cfgString)

	return &ConfigMock{
		AuthTokenFunc: func() (string, error) {
			return c.AuthToken()
		},
		GetFunc: func(s string) (string, error) {
			return c.Get(s)
		},
		SetFunc: func(s1 string, s2 string) {
			c.Set(s1, s2)
		},
		WriteFunc: func() error {
			return c.Write()
		},
	}
}

func StubWriteConfig(t *testing.T) func(io.Writer) {
	t.Helper()
	tempDir := t.TempDir()
	os.Setenv(JhConfigDir, tempDir)
	return func(w io.Writer) {
		config, err := os.Open(filepath.Join(tempDir, "config.yml"))
		if err != nil {
			return
		}
		defer config.Close()
		configData, err := io.ReadAll(config)
		if err != nil {
			return
		}
		_, err = w.Write(configData)
		if err != nil {
			return
		}
	}

}
