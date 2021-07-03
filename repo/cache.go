package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gitsight/go-vcsurl"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/model"
	"gorm.io/gorm"
)

/*
Git Repository Cache


*/

type Repository interface {
	Clone(context.Context) error // clone to path from URL
	Update(context.Context) error
	Size() (int64, error)
	// TODO: abstract away history, log and file access too
}

type GitRepository struct {
	Path string
	URL  string
}

type RepoCache struct {
	Base string
}

var (
	// ErrCacheExists indicates a different repository whith the same cache folder already exists
	ErrCacheExists = errors.New("cache folder for new repository already exists")
)

// Clone a git repository
func (g *GitRepository) Clone(ctx context.Context) error {
	r, err := git.PlainCloneContext(ctx, g.Path, false, &git.CloneOptions{
		URL:               g.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		return err
	}

	if err := r.FetchContext(ctx, &git.FetchOptions{}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	}

	return nil
}

// Update (fetch) a git repository
func (g *GitRepository) Update(ctx context.Context) error {
	r, err := git.PlainOpen(g.Path)
	if err != nil {
		return err
	}

	if err := r.FetchContext(ctx, nil); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	}

	return nil
}

func (c *RepoCache) urlToRepoPath(url string) (path string, err error) {
	v, err := vcsurl.Parse(url)
	if err != nil {
		return "", err
	}

	path = filepath.Join(c.Base, string(v.Host), v.Username, v.Name)
	return
}

// Add a new repository to the cache
func (c *RepoCache) Add(url string) (err error) {
	found, err := c.Has(url)
	if err != nil || found {
		return err
	}

	path, err := c.urlToRepoPath(url)
	if err != nil {
		return err
	}

	// create the cache directory
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return
			}
		} else {
			return err
		}
	} else {
		log.Error().Err(err).Msgf("repo cache folder conflict for %s, %s", url, path)
		return ErrCacheExists
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// TODO: cache different VCS types other than git
	repo := &GitRepository{URL: url, Path: path}
	if err = repo.Clone(ctx); err != nil {
		if ferr := os.RemoveAll(path); ferr != nil {
			// what a bad day :(
			log.Logger.Error().
				AnErr("rmdir", ferr).
				AnErr("git clone", err).
				Str("path", path).
				Msg("couldn't remove dir after failed clone")
		}
		if !errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return err
		}
	}

	r := &model.Repository{URL: url, Location: path, Type: "git"}
	result := model.DB.Create(r)
	return result.Error
}

// Has returns true if the repository is already cached
func (c *RepoCache) Has(url string) (found bool, err error) {
	r := &model.Repository{URL: url}

	result := model.DB.Model(r).Where(r).Find(r)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	} else if r.ID == uuid.Nil {
		return false, nil
	}

	return true, nil
}
