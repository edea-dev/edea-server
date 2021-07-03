package api

// SPDX-License-Identifier: EUPL-1.2

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/model"
)

// REST returns a handler for the given model
func REST(m API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Panic().Err(err).Msgf("could not parse form: %v", err)
		}

		v, err := m.DecodeForm(r.PostForm)
		if err != nil {
			log.Panic().Err(err).Msgf("could not decode: %v", err)
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
				log.Panic().Err(err).Msgf("could not get model: %v", err)
			}

			if err := json.NewEncoder(w).Encode(o); err != nil {
				log.Panic().Err(err).Msgf("could not encode model: %v", err)
			}
			return
		}

		// at this point we already know that the subject will be present
		// due to the middleware authenticating the request
		sub := r.Context().Value(model.AuthContextKey).(model.AuthClaims).Subject

		// delete an object, we don't need to validate it first
		if r.Method == http.MethodDelete {
			if err := m.Delete(v, sub); err != nil {
				log.Panic().Err(err).Msgf("could not update: %v", err)
			}
			return
		}

		if err := m.Validate(v); err != nil {
			log.Panic().Err(err).Msgf("validation error: %v", err)
		}

		// update an object
		if r.Method == http.MethodPut {
			if _, err := m.Update(v, sub); err != nil {
				log.Panic().Err(err).Msgf("could not update: %v", err)
			}
		} else if r.Method == http.MethodPost { // create a new object
			o, err := m.Create(v, sub)
			if err != nil {
				log.Panic().Err(err).Msgf("could not create: %v", err)
			}

			// encode and send new object as json to the user
			if err := json.NewEncoder(w).Encode(o); err != nil {
				log.Panic().Err(err).Msgf("could not encode model: %v", err)
			}
		} else {
			log.Panic().Msgf("unsupported method: %s", r.Method)
		}
	}
}
