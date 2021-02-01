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

var (
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

		claims := model.AuthClaims{}
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
		result := model.DB.Model(user).Where(user).First(user)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("could not fetch user data for %s", claims.Subject)
		}

		// add claims and user object to the context
		ctx := context.WithValue(r.Context(), model.AuthContextKey, claims)
		ctx = context.WithValue(ctx, util.UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// InitJWKS unmarshals the keyset
func InitJWKS(doc []byte) error {
	return json.Unmarshal(doc, keySet)
}

func createUserIfNotExist(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(model.AuthContextKey).(model.AuthClaims)
	if model.UserExists(claims.Subject) {
		log.Info().Msg("user already exists")
		return
	}

	// create a new user in the database if they're logging in for the first time

	u := model.User{
		AuthUUID: claims.Subject,
		Handle:   claims.Subject, // TODO: figure that one out. we need to handle that better (handle, get it?)
	}

	if result := model.DB.Model(&u).Create(&u); result.Error != nil {
		log.Error().Str("auth_uuid", claims.Subject).Err(result.Error).Msgf("could not create new user")
	}

	p := model.Profile{DisplayName: claims.Nickname, Avatar: claims.Picture, UserID: u.ID}

	if result := model.DB.Model(&p).Create(&p); result.Error != nil {
		log.Panic().Str("auth_uuid", claims.Subject).Err(result.Error).Msgf("could not create new profile")
	}

	log.Info().EmbedObject(&u).Msgf("created a new user")
}
