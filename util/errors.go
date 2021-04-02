package util

// SPDX-License-Identifier: EUPL-1.2

import "errors"

var (
	// ErrImSorryDave is returned when a user tries to manipulate form data
	ErrImSorryDave = errors.New("i'm sorry dave, i'm afraid i can't do that")

	// ErrNoSuchBench is returned if a user tries to modify a bench which doesn't exist or doesn't belong to them
	ErrNoSuchBench = errors.New("no such bench")

	// ErrNoSuchModule is returned if a module does not exist (anymore)
	ErrNoSuchModule = errors.New("no such module")

	// ErrNoActiveBench is returned if there is no bench currently set to active
	ErrNoActiveBench = errors.New("no currently active bench")
)
