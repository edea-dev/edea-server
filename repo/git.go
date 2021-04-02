package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

type Git struct {
	URL string
}

// Readme searches for a readme.md file in the repository and returns it if found
func (g *Git) Readme() (string, error) {
	return g.File("readme.md", false)
}

// File searches for a given file in the git respository
func (g *Git) File(name string, caseSensitive bool) (string, error) {
	if found, err := cache.Has(g.URL); err != nil {
		return "", err
	} else if !found {
		return "", ErrUncachedRepo
	}

	if cache == nil {
		return "", fmt.Errorf("cache is not initialized")
	}

	path, err := cache.urlToRepoPath(g.URL)
	if err != nil {
		return "", err
	}

	r, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}

	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	if err != nil {
		return "", err
	}

	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return "", err
	}

	// ... retrieve the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return "", err
	}

	var file *object.File

	// ... get the files iterator and print the file
	tree.Files().ForEach(func(f *object.File) error {
		if caseSensitive {
			if f.Name == name {
				file = f
				return storer.ErrStop
			}
		}
		if strings.ToLower(f.Name) == name {
			file = f
			return storer.ErrStop
		}
		return nil
	})

	if file != nil {
		bin, err := file.IsBinary()
		if bin {
			return "", ErrNoFile
		} else if err != nil {
			return "", err
		}
		s, err := file.Contents()
		if err != nil {
			return "", err
		}
		return s, nil
	}

	return "", ErrNoFile
}
