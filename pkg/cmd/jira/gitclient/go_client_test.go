package gitclient

import (
	"bytes"
	"testing"

	"github.com/stirboy/jh/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestCreateBranchWithCheckout(t *testing.T) {
	// given
	repo := StubGitRepository(t)
	out := &bytes.Buffer{}
	io := &iostreams.IOStream{
		Out: out,
	}
	c := NewClient(repo, io)

	// when
	err := c.CreateBranchWithCheckout("feature/test")

	// then
	assert.NoError(t, err)
	assert.Equal(t, "switched to branch: 'feature/test'\n", out.String())
}
