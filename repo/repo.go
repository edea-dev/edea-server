package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/model"
	"gitlab.com/edea-dev/edead/util"
	"gopkg.in/yaml.v2"
)

type Info struct {
	LastCommit struct {
		Time time.Time
		Hash string
	}
	Readme string
}

var (
	ErrExists             = errors.New("repository already added")
	ErrNoFile             = errors.New("no file found")
	ErrBadCredentials     = errors.New("bad credentials")
	ErrUnexpectedResponse = errors.New("unexpected http response")
	ErrUncachedRepo       = errors.New("repository not cached")

	cache *RepoCache
)

func New(url string) error {
	found, err := cache.Has(url)
	if err != nil {
		return err
	}
	if found {
		return ErrExists
	}

	return cache.Add(url)
}

func Add(url string) error {
	return cache.Add(url)
}

func InitCache(path string) {
	cache = &RepoCache{Base: path}
}

func GetModulePath(mod *model.Module) (string, error) {
	g := &Git{URL: mod.RepoURL}
	p := &Project{}

	// read and parse the module configuration out of the repo
	s, err := g.File("edea.yml", false)
	if err != nil {
		// assuming old format, i.e. no sub-modules
		log.Info().Msgf("module %s does not contain an edea.yml file, assuming project files are in top-level dir", mod.ID)

		repoDir, _ := g.Dir()
		return repoDir, nil
	}
	if err := yaml.Unmarshal([]byte(s), p); err != nil {
		return "", util.HintError{
			Hint: fmt.Sprintf("Could not parse edea.yml for \"%s\"\nTry checking if the syntax is correct.", mod.Name),
			Err:  err,
		}
	}

	v, ok := p.Modules[mod.Sub]
	if !ok {
		log.Panic().Err(errors.New("sub-module specified but does not exist")).Msg("the sub-module key in the database does not exist in the repo edea.yml")
	}

	repoDir, _ := g.Dir() // at this point we already know the it's cached
	dir := strings.ReplaceAll(v.Directory, "../", "")
	dir = strings.TrimPrefix(dir, "/")
	dir = filepath.Join(repoDir, dir)
	return dir, nil
}
