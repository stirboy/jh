package gitclient

func NewGitClientMock() *GitClientMock {
	return &GitClientMock{
		CreateBranchWithCheckoutFunc: func(s string) error {
			return nil
		},
	}
}
