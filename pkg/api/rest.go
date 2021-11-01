package api

// SPDX-License-Identifier: EUPL-1.2

import (
	"encoding/json"
	"net/http"

	"gitlab.com/edea-dev/edead/internal/model"
	"go.uber.org/zap"
)

// REST returns a handler for the given model
func REST(m API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			zap.L().Panic("could not parse form", zap.Error(err))
		}

		v, err := m.DecodeForm(r.PostForm)
		if err != nil {
			zap.L().Panic("could not decode", zap.Error(err))
		}

		// GET request
		if r.Method == "" || r.Method == http.MethodGet {
			var sub string

			// some get methods require you to be logged in, check if we have a subject in the context
			subValue := r.Context().Value(model.AuthContextKey)
			if subValue != nil {
				sub = subValue.(model.AuthClaims).Subject
			}

			o, err := m.Get(v, sub)

			if err != nil {
				zap.L().Panic("could not get model", zap.Error(err))
			}

			if err := json.NewEncoder(w).Encode(o); err != nil {
				zap.L().Panic("could not encode model", zap.Error(err))
			}
			return
		}

		// at this point we already know that the subject will be present
		// due to the middleware authenticating the request
		sub := r.Context().Value(model.AuthContextKey).(model.AuthClaims).Subject

		// delete an object, we don't need to validate it first
		if r.Method == http.MethodDelete {
			if err := m.Delete(v, sub); err != nil {
				zap.L().Panic("could not update", zap.Error(err))
			}
			return
		}

		if err := m.Validate(v); err != nil {
			zap.L().Panic("validation error", zap.Error(err))
		}

		// update an object
		if r.Method == http.MethodPut {
			if _, err := m.Update(v, sub); err != nil {
				zap.L().Panic("could not update", zap.Error(err))
			}
		} else if r.Method == http.MethodPost { // create a new object
			o, err := m.Create(v, sub)
			if err != nil {
				zap.L().Panic("could not create", zap.Error(err))
			}

			// encode and send new object as json to the user
			if err := json.NewEncoder(w).Encode(o); err != nil {
				zap.L().Panic("could not encode model", zap.Error(err))
			}
		} else {
			zap.L().Panic("unsupported method", zap.String("method", r.Method))
		}
	}
}
