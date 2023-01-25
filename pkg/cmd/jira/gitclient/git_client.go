package gitclient

import (
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stirboy/jh/pkg/iostreams"
)

//go:generate moq -rm -out git_client_mock.go . GitClient
type GitClient interface {
	CreateBranchWithCheckout(string) error
}

// client implements GitClient
type Client struct {
	GitPath string

	Stdout io.Writer
}

func NewClient(path string, iostream *iostreams.IOStream) GitClient {
	return &Client{
		GitPath: path,
		Stdout:  iostream.Out,
	}
}

// CreateBranchWithCheckout creates a new branch in current directory
// and does a checkout to that branch. All local changes will be preserved.
// If current directory is not a git repo, error is returned
func (c *Client) CreateBranchWithCheckout(branchName string) error {
	r, err := git.PlainOpen(c.GitPath)
	if err != nil {
		return fmt.Errorf("jh create branch failed: %w", err)
	}

	worktree, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("jh create branch failed: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
		Keep:   true,
		
	})
	if err != nil {
		return fmt.Errorf("jh create branch failed: %w", err)
	}

	fmt.Fprintf(c.Stdout, "switched to branch: '%v'\n", branchName)

	return nil
}
