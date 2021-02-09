package repo

import (
	"errors"
	"time"
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

func InitCache(path string) {
	cache = &RepoCache{Base: path}
}
