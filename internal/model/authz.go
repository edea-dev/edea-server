package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
)

// AuthClaims used
type AuthClaims struct {
	Subject  string `json:"sub,omitempty"`
	Picture  string `json:"picture,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

var (
	// ErrUnauthorized is returned when a user is neither an admin nor an owner of the model to be changed
	ErrUnauthorized = errors.New("user is not authorized to change this row")
	// ErrNoSuchSubject is returned on empty sub parameter or if no user with a matching subject exists
	ErrNoSuchSubject = errors.New("no subject given or subject does not exist")
)

func isInUsers(users []*User, sub string) bool {
	for _, u := range users {
		if u.AuthUUID == sub {
			return true
		}
	}

	return false
}
