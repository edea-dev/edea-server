package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"testing"
)

func Test_Git_Readme_Success(t *testing.T) {
	if err := cache.Add("https://github.com/tachiniererin/nargh"); err != nil {
		t.Fatal(err)
	}

	r := Git{URL: "https://github.com/tachiniererin/nargh"}
	if s, err := r.Readme(); err != nil || s == "" {
		t.Error(err)
	}
}

/* TODO: fixme
func Test_Git_Readme_EmptyRepoFail(t *testing.T) {
	if err := cache.Add("https://gitlab.com/tachiniererin/empty"); err != nil {
		t.Fatal(err)
	}

	r := Git{URL: "https://gitlab.com/tachiniererin/empty"}
	_, err := r.Readme()
	if err != nil {
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return
		}
		t.Error(err)
	} else {
		t.Fatal("expected an error")
	}
	t.Fail()
}
*/
func Test_Git_Readme_NoReadme(t *testing.T) {
	if err := cache.Add("https://github.com/tachiniererin/ma-updater"); err != nil {
		t.Fatal(err)
	}

	r := Git{URL: "https://github.com/tachiniererin/ma-updater"}
	_, err := r.Readme()
	if err != nil {
		if errors.Is(err, ErrNoFile) {
			return
		}
		t.Error(err)
	} else {
		t.Fatal("expected an error")
	}
	t.Fail()
}
