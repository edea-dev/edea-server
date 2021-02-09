package repo

import (
	"errors"
	"testing"
)

func TestGitLab_GetReadmeSuccess(t *testing.T) {
	g := GitLab{
		FullPath: "gitlab-org/libgit2",
		Host:     "gitlab.com",
	}

	s, err := g.Readme()
	if err != nil {
		t.Fatal(err)
		return
	}

	if s == "" {
		t.Fatalf("empty readme")
		return
	}
}

func TestGitLab_GetReadmeFail_NoReadme(t *testing.T) {
	g := GitLab{
		FullPath: "tachiniererin/empty",
		Host:     "gitlab.com",
	}

	_, err := g.Readme()
	if err == nil {
		t.Fatal("expected an error")
	}
	if err != nil && !errors.Is(err, ErrNoFile) {
		t.Fatal(err)
		return
	}
}

/*
func TestGitLab_Error_InvalidAuthToken(t *testing.T) {
	g := GitLab{
		FullPath: "gitlab-org/libgit2",
		Host:     "gitlab.com",
	}

	_, _, err := g.RateLimit()
	if err == nil {
		t.Fatal(err)
		return
	}
	if e, ok := err.(*ErrGitLabGQL); ok {
		if e.Msg != "Bad credentials" {
			t.Fatal(e)
		}
	} else {
		t.Fatalf("unexpected error: %v", err)
	}
}
*/
