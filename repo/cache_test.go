package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

func TestRepo_AddToCache(t *testing.T) {
	r := RepoCache{Base: "./_tmp/git"}
	err := r.Add("https://github.com/tachiniererin/nargh")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRepo_AddToCache_FailNotExist(t *testing.T) {
	r := RepoCache{Base: "./_tmp/git"}
	err := r.Add("https://github.com/tachiniererin/narg")
	if err != nil {
		if errors.Is(err, transport.ErrAuthenticationRequired) {
			return
		}
		t.Fatal(err)
	}
	t.Fail()
}
