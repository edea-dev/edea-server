package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/rs/zerolog/log"
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
	Doc       string `yaml:"doc"`
	// TODO: add configuration here
}

type Commit struct {
	Message string
	Ref     string
}

// Readme searches for a readme.md file in the repository and returns it if found
func (g *Git) Readme() (string, error) {
	return g.File("readme.md", false)
}

// SubModuleDir looks if a sub-module exists or else returns the base path
func (g *Git) SubModuleDir(sub string) (string, error) {
	p := &Project{}

	// read and parse the module configuration out of the repo
	s, err := g.File("edea.yml", false)
	if err != nil {
		return "", errors.New("module does not contain an edea.yml file")
	}
	if err := yaml.Unmarshal([]byte(s), p); err != nil {
		return "", err
	}

	var path string

	m, ok := p.Modules[sub]
	if !ok {
		path = m.Directory
	} else {
		path = filepath.Base(m.Directory)
	}

	return path, nil
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
		path := filepath.Join(filepath.Base(m.Directory), filepath.Base(m.Readme))
		return g.File(path, true)
	}

	return g.File(filepath.Join(filepath.Base(m.Directory), "readme.md"), false)
}

// HasDocs searches for a book.toml file in the repository and returns true if found
func (g *Git) HasDocs(sub string) (bool, error) {
	p := &Project{}

	// read and parse the module configuration out of the repo
	s, err := g.File("edea.yml", false)
	if err != nil {
		return false, errors.New("module does not contain an edea.yml file")
	}
	if err := yaml.Unmarshal([]byte(s), p); err != nil {
		return false, err
	}

	var path string

	m, ok := p.Modules[sub]
	if !ok {
		path = "book.toml"
	} else {
		if m.Doc != "" {
			path = filepath.Join(filepath.Base(m.Directory), filepath.Base(m.Doc), "book.toml")
		} else {
			path = filepath.Join(filepath.Base(m.Directory), "book.toml")
		}
	}

	log.Info().Msgf("book.toml path: %s", path)

	book, err := g.File(path, false)
	if err != nil {
		if errors.Is(err, ErrNoFile) {
			return false, nil
		}
		return false, err
	}
	if len(book) > 0 {
		return true, nil
	}
	return false, fmt.Errorf("empty book.toml found")
}

// SubModuleDocs searches for the doc subfolder in the module and returns all .md files
func (g *Git) SubModuleDocs(sub string) (string, error) {
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
	if m.Doc != "" {
		return filepath.Join(filepath.Base(m.Directory), filepath.Base(m.Doc)), nil
	}

	return filepath.Base(m.Directory), nil
}

// File searches for a given file in the git respository
func (g *Git) File(name string, caseSensitive bool) (string, error) {
	// ... retrieves the branch pointed by HEAD
	r, ref, err := g.head()
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
	r, err := g.open()
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	if err := w.Pull(&git.PullOptions{RemoteName: "origin"}); err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

func (g *Git) open() (*git.Repository, error) {
	if found, err := cache.Has(g.URL); err != nil {
		return nil, err
	} else if !found {
		return nil, ErrUncachedRepo
	}

	if cache == nil {
		return nil, fmt.Errorf("cache is not initialized")
	}

	path, err := cache.urlToRepoPath(g.URL)
	if err != nil {
		return nil, err
	}

	return git.PlainOpen(path)
}

func (g *Git) head() (*git.Repository, *plumbing.Reference, error) {
	r, err := g.open()
	if err != nil {
		return nil, nil, err
	}

	// retrieves the branch pointed by HEAD
	ref, err := r.Head()
	return r, ref, err
}

// History returns the commits and the reference hash for a repository or submodule
func (g *Git) History(folder string) ([]*Commit, error) {
	r, ref, err := g.head()
	if err != nil {
		return nil, err
	}

	since := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Now()
	options := &git.LogOptions{From: ref.Hash(), Since: &since, Until: &until}
	if folder != "" {
		options.PathFilter = func(path string) bool {
			return strings.HasPrefix(path, folder)
		}
	}
	var commits []*Commit

	cIter, err := r.Log(options)
	if err != nil {
		log.Panic().Err(err).Msg("could not retrieve history of repo")
	}
	err = cIter.ForEach(func(c *object.Commit) error {
		msg := strings.ReplaceAll(c.String(), "\n", "<br>")
		v := &Commit{Message: msg, Ref: c.Hash.String()}
		commits = append(commits, v)

		fmt.Println(c.String())

		return nil
	})

	return commits, err
}

// FileAt searches for a given file with the specific revision in the git respository
// the revision parameter can be anything ResolveRevision understands (tags, branches, HEAD^1, etc.)
func (g *Git) FileAt(name string, caseSensitive bool, revision string) ([]byte, error) {
	// ... retrieves the branch pointed by HEAD
	r, err := g.open()
	if err != nil {
		return nil, err
	}

	ref, err := r.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return nil, err
	}

	// ... retrieving the commit object
	commit, err := r.CommitObject(*ref)
	if err != nil {
		return nil, err
	}

	// ... retrieve the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
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
		s, err := file.Contents()
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	}

	return nil, ErrNoFile
}

// FileByExtAt searches for a given file by extension with the specific revision in the git respository
// the revision parameter can be anything ResolveRevision understands (tags, branches, HEAD^1, etc.)
// NOTE: ext *must* contain the . (dot), e.g. ".kicad_pcb"
func (g *Git) FileByExtAt(dir, ext, revision string) ([]byte, error) {
	// ... retrieves the branch pointed by HEAD
	r, err := g.open()
	if err != nil {
		return nil, err
	}

	ref, err := r.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return nil, err
	}

	// ... retrieving the commit object
	commit, err := r.CommitObject(*ref)
	if err != nil {
		return nil, err
	}

	// ... retrieve the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var file *object.File

	// find the first file by extension
	tree.Files().ForEach(func(f *object.File) error {
		if filepath.Dir(f.Name) == dir && filepath.Ext(f.Name) == ext {
			file = f
			return storer.ErrStop
		}
		return nil
	})

	if file != nil {
		s, err := file.Contents()
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	}

	return nil, ErrNoFile
}
