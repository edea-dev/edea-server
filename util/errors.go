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

// HintError wraps an error with a descriptive hint for the user
type HintError struct {
	Hint string
	Err  error
}

func (e HintError) Error() string {
	return e.Hint + ": " + e.Err.Error()
}

// Unwrap the error
func (e HintError) Unwrap() error {
	return e.Err
}
