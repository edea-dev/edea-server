package view

// SPDX-License-Identifier: EUPL-1.2

import (
	"fmt"
	"net/http"

	"gitlab.com/edea-dev/edea/backend/api"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
)

// CurrentUser returns the full User object when logged in or nil otherwise
func CurrentUser(r *http.Request) *model.User {
	u, ok := r.Context().Value(util.UserContextKey).(*model.User)
	if !ok {
		return nil
	}

	return u
}

// CurrentSubject fetches the JWT subject from the request context or returns an empty string if the request is unauthenticated
func CurrentSubject(r *http.Request) string {
	claimsValue := r.Context().Value(model.AuthContextKey)
	if claimsValue == nil {
		return ""
	}

	claims := claimsValue.(model.AuthClaims)
	return claims.Subject
}

func decodeAndValidate(r *http.Request, m api.API) (interface{}, string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, "", fmt.Errorf("could not parse form: %w", err)
	}

	v, err := m.DecodeForm(r.PostForm)
	if err != nil {
		return nil, "", fmt.Errorf("could not decode model: %w", err)
	}

	if err := m.Validate(v); err != nil {
		return nil, "", fmt.Errorf("could not validate model: %w", err)
	}

	return v, CurrentSubject(r), nil
}

// CreateModel parses, decodes, validates and creates a model
func CreateModel(r *http.Request, m api.API) (interface{}, error) {
	v, sub, err := decodeAndValidate(r, m)
	if err != nil {
		return nil, err
	}

	o, err := m.Create(v, sub)
	if err != nil {
		return nil, fmt.Errorf("could not create new model: %w", err)
	}
	return o, nil
}

// UpdateModel parses, decodes, validates and updates a model
func UpdateModel(r *http.Request, m api.API) (interface{}, error) {
	v, sub, err := decodeAndValidate(r, m)
	if err != nil {
		return nil, err
	}

	o, err := m.Update(v, sub)
	if err != nil {
		return nil, fmt.Errorf("could not update model: %w", err)
	}
	return o, nil
}

// GetModel parses, decodes, validates and fetches a model
func GetModel(r *http.Request, m api.API) (interface{}, error) {
	v, sub, err := decodeAndValidate(r, m)
	if err != nil {
		return nil, err
	}

	o, err := m.Get(v, sub)
	if err != nil {
		return nil, fmt.Errorf("could not update model: %w", err)
	}
	return o, nil
}
