package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"gopkg.in/yaml.v2"
)

type Git struct {
	URL string
}

// Project is the top level project configuration
type Project struct {
	Name    string            `yaml:"name"`
	Modules map[string]Module `yaml:"modules"`
}

// Module references the schematic and pcb for this module
type Module struct {
	Readme    string `yaml:"readme"`
	Directory string `yaml:"dir"`
	// TODO: add configuration here
}

// Readme searches for a readme.md file in the repository and returns it if found
func (g *Git) Readme() (string, error) {
	return g.File("readme.md", false)
}

// SubModuleReadme searches for a readme.md file in the repository and returns it if found
func (g *Git) SubModuleReadme(sub string) (string, error) {
	p := &Project{}

	// read and parse the module configuration out of the repo
	s, err := g.File("edea.yml", false)
	if err != nil {
		return "", errors.New("module does not contain an edea.yml file")
	}
	if err := yaml.Unmarshal([]byte(s), p); err != nil {
		return "", err
	}

	m, ok := p.Modules[sub]
	if !ok {
		return "", errors.New("no such sub-module")
	}

	// sanitise the filepaths a bit, we only expect single level nesting
	// if the git library already does it, we could skip this, but needs verification
	if m.Readme != "" {
		return g.File(filepath.Join(filepath.Base(m.Directory), filepath.Base(m.Readme)), false)
	}

	return g.File(filepath.Join(filepath.Base(m.Directory), "readme.md"), false)
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

// Dir returns the location of the checked out repository
func (g *Git) Dir() (string, error) {
	if found, err := cache.Has(g.URL); err != nil {
		return "", err
	} else if !found {
		return "", ErrUncachedRepo
	}

	if cache == nil {
		return "", fmt.Errorf("cache is not initialized")
	}

	return cache.urlToRepoPath(g.URL)
}

// Pull the latest changes from the origin
func (g *Git) Pull() error {
	if found, err := cache.Has(g.URL); err != nil {
		return err
	} else if !found {
		return ErrUncachedRepo
	}

	if cache == nil {
		return fmt.Errorf("cache is not initialized")
	}

	path, err := cache.urlToRepoPath(g.URL)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	if err := w.Pull(&git.PullOptions{RemoteName: "origin"}); err == git.NoErrAlreadyUpToDate {
		return nil
	} else {
		return err
	}
}
