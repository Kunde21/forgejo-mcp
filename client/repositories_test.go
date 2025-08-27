package client

import (
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-cmp/cmp"
)

func TestTransformRepository(t *testing.T) {
	giteaRepo := &gitea.Repository{
		ID:          1,
		Name:        "test-repo",
		FullName:    "owner/test-repo",
		Description: "A test repository",
		HTMLURL:     "https://example.com/owner/test-repo",
		CloneURL:    "https://example.com/owner/test-repo.git",
	}

	want := Repository{
		ID:          1,
		Name:        "test-repo",
		FullName:    "owner/test-repo",
		Description: "A test repository",
		HTMLURL:     "https://example.com/owner/test-repo",
		CloneURL:    "https://example.com/owner/test-repo.git",
	}
	got := transformRepository(giteaRepo)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

	gotNil := transformRepository(nil)
	if !cmp.Equal(Repository{}, gotNil) {
		t.Error(cmp.Diff(Repository{}, gotNil))
	}
}
