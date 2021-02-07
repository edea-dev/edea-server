package view

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/api"
	"gitlab.com/edea-dev/edea/backend/model"
)

// CurrentUser returns the full User object when logged in or nil otherwise
func CurrentUser(r *http.Request) *model.User {
	claims, ok := r.Context().Value(model.AuthContextKey).(model.AuthClaims)
	if !ok {
		return nil
	}

	u := &model.User{AuthUUID: claims.Subject}

	if result := model.DB.Where(u).First(u); result.Error != nil {
		log.Error().Err(result.Error).Msgf("could not fetch user data for %s", claims.Subject)
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
