package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"gitlab.com/edea-dev/edea-server/internal/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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
	Readme    string `yaml:"readme"` // readme file or folder which contains readme.md
	Directory string `yaml:"dir"`    // path to the kicad project file or folder which contains it
	Doc       string `yaml:"doc"`    // path to book.toml
	// TODO: add configuration here
}

type Commit struct {
	Message string
	Ref     string
}

// Readme searches for a readme.md file in the repository and returns it if found
func (g *Git) Readme(revision string) (string, error) {
	b, err := g.FileAt("readme.md", false, revision)
	return string(b), err
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
func (g *Git) SubModuleReadme(sub, revision string) (string, error) {
	p := &Project{}

	// read and parse the module configuration out of the repo
	s, err := g.FileAt("edea.yml", false, revision)
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
		b, err := g.FileAt(path, true, revision)
		return string(b), err
	}
	b, err := g.FileAt(filepath.Join(filepath.Base(m.Directory), "readme.md"), false, revision)
	return string(b), err
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
		zap.L().Panic("could not retrieve history of repo", zap.Error(err))
	}
	err = cIter.ForEach(func(c *object.Commit) error {
		msg := strings.ReplaceAll(c.String(), "\n", "<br>")
		v := &Commit{Message: msg, Ref: c.Hash.String()}
		commits = append(commits, v)

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

type GitFile struct {
	Name    string
	Content []byte
}

// FilesByExtAt searches for a given file by extension with the specific revision in the git respository
// the revision parameter can be anything ResolveRevision understands (tags, branches, HEAD^1, etc.)
// NOTE: ext *must* contain the . (dot), e.g. ".kicad_pcb"
func (g *Git) FilesByExtAt(dir, ext, revision string) (files []GitFile, err error) {
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

	// find the first file by extension
	err = tree.Files().ForEach(func(f *object.File) error {
		if filepath.Dir(f.Name) == dir && filepath.Ext(f.Name) == ext {
			s, err := f.Contents()
			if err != nil {
				return err
			}
			files = append(files, GitFile{filepath.Base(f.Name), []byte(s)})
			return storer.ErrStop
		}
		return nil
	})

	return
}

// FileByExtAt works like FilesByExtAt but only returns the first file found
func (g *Git) FileByExtAt(dir, ext, revision string) ([]byte, error) {
	files, err := g.FilesByExtAt(dir, ext, revision)
	if len(files) > 0 {
		return files[0].Content, err
	}
	return nil, err
}

// SchematicHelper pulls all the schematic files and the symbol cache from the repository
// at the specified revision and copies them to a temporary folder
func (g *Git) SchematicHelper(dir, revision string) (map[string]string, error) {
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

	dest, err := os.MkdirTemp("", "*")
	if err != nil {
		return nil, err
	}

	// clean up afterwards
	defer os.RemoveAll(dest)

	var libFile string
	var sch []string

	// find the first file by extension
	err = tree.Files().ForEach(func(f *object.File) error {
		if filepath.Dir(f.Name) == dir {
			if strings.HasSuffix(filepath.Base(f.Name), "-cache.lib") {
				libFile, err = gitFileToTemp(f, dest)
				if err != nil {
					return err
				}
			} else if filepath.Ext(f.Name) == ".sch" {
				fn, err := gitFileToTemp(f, dest)
				sch = append(sch, fn)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if libFile == "" {
		return nil, fmt.Errorf("no -cache.lib file was found")
	}

	svgs := make(map[string]string)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, fn := range sch {
		argv := []string{"-l", libFile, "-f", fn}

		plotCmd := exec.CommandContext(ctx, "plotkicadsch", argv...)

		// run the plotting operation
		logOutput, err := plotCmd.CombinedOutput()
		if err != nil {
			return nil, util.HintError{
				Hint: fmt.Sprintf("could not plot %s:\n%s", fn, string(logOutput)),
				Err:  err,
			}
		}

		svgName := fmt.Sprintf("%s.svg", strings.TrimSuffix(filepath.Base(fn), filepath.Ext(fn)))
		svg := filepath.Join(filepath.Dir(fn), svgName)

		cleanerCmd := exec.CommandContext(ctx, "svgcleaner", svg, svg)
		out, err := cleanerCmd.CombinedOutput()
		if err != nil {
			zap.L().Error("could not run svgcleaner", zap.ByteString("output", out))
		}

		b, err := os.ReadFile(svg)
		if err != nil {
			return nil, util.HintError{
				Hint: fmt.Sprintf("could not read svg %s", svg),
				Err:  err,
			}
		}
		svgs[svgName] = string(b)
	}

	return svgs, err
}

func gitFileToTemp(f *object.File, dest string) (string, error) {
	tf, err := os.Create(filepath.Join(dest, filepath.Base(f.Name)))
	if err != nil {
		return "", err
	}
	src, err := f.Reader()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(tf, src)
	if err != nil {
		return "", err
	}
	name := tf.Name()
	return name, tf.Close()
}
