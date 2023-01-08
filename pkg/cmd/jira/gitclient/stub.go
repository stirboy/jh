package gitclient

import (
	"testing"

	"github.com/go-git/go-git/v5"
)

func NewGitClientMock() *GitClientMock {
	return &GitClientMock{
		CreateBranchWithCheckoutFunc: func(s string) error {
			return nil
		},
	}
}

func StubGitRepository(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	_, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL: "https://github.com/stirboy/jh",
	})
	if err != nil {
		t.Error(err)
	}

	return tempDir
}
