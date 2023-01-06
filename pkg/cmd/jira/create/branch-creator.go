package create

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func CreateGitBranch(issueKey string, ops *CreateOptions) error {
	// create and checkout to new branch
	if ops.CreateGitBranch == "" {
		return nil
	}

	// printing out a line for formatting
	fmt.Println()

	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("jh create branch failed: %w", err)
	}

	r, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("jh create branch failed: %w", err)
	}

	worktree, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("jh create branch failed: %w", err)
	}

	branchName := strings.Replace(ops.CreateGitBranch, "@", strings.ToLower(issueKey), 1)
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
		Keep:   true,
	})
	if err != nil {
		return fmt.Errorf("jh create branch failed: %w", err)
	}

	fmt.Printf("switched to branch: '%v'\n", branchName)

	return nil
}
