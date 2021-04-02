package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"testing"
	"time"
)

func TestGitHub_GetReadmeSuccess(t *testing.T) {
	g := GitHub{
		ProjectName:  "nargh",
		ProjectOwner: "tachiniererin",
		Host:         "github.com",
		AuthToken:    cfg.API.GitHubToken,
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

func TestGitHub_GetReadmeFail_NoReadme(t *testing.T) {
	g := GitHub{
		ProjectName:  "ma-updater",
		ProjectOwner: "tachiniererin",
		Host:         "github.com",
		AuthToken:    cfg.API.GitHubToken,
	}

	_, err := g.Readme()
	if err != nil && !errors.Is(err, ErrNoFile) {
		t.Fatal(err)
		return
	}
}

func TestGitHub_Error_InvalidAuthToken(t *testing.T) {
	g := GitHub{
		ProjectName:  "nargh",
		ProjectOwner: "tachiniererin",
		Host:         "github.com",
		AuthToken:    "shibboleth",
	}

	_, _, err := g.RateLimit()
	if err == nil {
		t.Fatal(err)
		return
	}
	if e, ok := err.(*GraphQLResponse); ok {
		if e.Message != "Bad credentials" {
			t.Fatal(e)
		}
	} else {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestGitHub_GetRateLimit(t *testing.T) {
	g := GitHub{
		ProjectName:  "nargh",
		ProjectOwner: "tachiniererin",
		Host:         "github.com",
		AuthToken:    cfg.API.GitHubToken,
	}

	_, tm, err := g.RateLimit()
	if err != nil {
		t.Fatal(err)
		return
	}

	if tm.Add(time.Hour).Before(time.Now().UTC()) {
		t.Fatal("rate limit should reset every hour")
	}
}
