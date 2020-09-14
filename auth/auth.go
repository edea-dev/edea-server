package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/square/go-jose/jwt"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gopkg.in/square/go-jose.v2"
)

// Provider interface to be implemented by Identity Providers
type Provider interface {
	CallbackHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
}

type Claims struct {
	Subject  string `json:"sub,omitempty"`
	Picture  string `json:"picture,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

var (
	// ContextKey is the request context key for the session data
	ContextKey = Claims{}

	keySet *jose.JSONWebKeySet
)

// Middleware checks if there is a valid json web token in the request
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var raw string

		auth := r.Header.Get("Authorization")
		s, err := r.Cookie("jwt")

		if err != nil && len(auth) == 0 {
			log.Debug().Msg("unauthenticated request")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Authorization header/session cookie missing"))
			return
		}

		if len(auth) > 0 {
			raw = strings.Replace(auth, "Bearer ", "", 1)
		} else {
			raw = s.Value
		}

		tok, err := jwt.ParseSigned(raw)
		if err != nil {
			log.Panic().Err(err).Msg("could not parse jwt")
		}

		claims := Claims{}
		scopes := struct {
			Scopes []string
		}{}

		// verify claims
		if err := tok.Claims(keySet, &claims, &scopes); err != nil {
			log.Error().Err(err).Msgf("could not verify jwt")

			// remove offending jwt cookie
			cookie := http.Cookie{
				Name:     "jwt",
				Value:    "",
				Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, &cookie)

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Error verifying JWT token: " + err.Error()))
			return
		}

		// get the current user object from the database
		user := &model.User{AuthUUID: claims.Subject}
		if err = model.DB.Model(user).Where("auth_uuid = ?", claims.Subject).Select(); err != nil {
			log.Error().Err(err).Msgf("could not fetch user data for %s", claims.Subject)
		}

		// add claims and user object to the context
		ctx := context.WithValue(r.Context(), ContextKey, claims)
		ctx = context.WithValue(ctx, util.UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// InitJWKS unmarshals the keyset
func InitJWKS(doc []byte) error {
	return json.Unmarshal(doc, keySet)
}

func createUserIfNotExist(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(ContextKey).(Claims)
	if model.UserExists(claims.Subject) {
		log.Info().Msg("user already exists")
		return
	}

	// create a new user in the database if they're logging in for the first time

	p := model.Profile{DisplayName: claims.Nickname, Avatar: claims.Picture}

	if _, err := model.DB.Model(&p).Insert(); err != nil {
		log.Panic().Str("auth_uuid", claims.Subject).Err(err).Msgf("could not create new profile")
	}

	u := model.User{
		AuthUUID:  claims.Subject,
		Handle:    claims.Subject, // TODO: figure that one out. we need to handle that better (handle, get it?)
		ProfileID: p.ID,
	}

	if _, err := model.DB.Model(&u).Insert(); err != nil {
		log.Error().Str("auth_uuid", claims.Subject).Err(err).Msgf("could not create new user")
	}

	log.Info().EmbedObject(&u).Msgf("created a new user")
}
