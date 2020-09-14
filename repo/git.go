package repo

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

func (g *Git) Readme() (string, error) {
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

	var readme *object.File

	// ... get the files iterator and print the file
	tree.Files().ForEach(func(f *object.File) error {
		if strings.ToLower(f.Name) == "readme.md" {
			readme = f
			return storer.ErrStop
		}
		return nil
	})

	if readme != nil {
		bin, err := readme.IsBinary()
		if bin {
			return "", ErrNoReadme
		} else if err != nil {
			return "", err
		}
		s, err := readme.Contents()
		if err != nil {
			return "", err
		}
		return s, nil
	}

	return "", ErrNoReadme
}
