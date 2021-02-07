package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
)

// Provider interface to be implemented by Identity Providers
type Provider interface {
	CallbackHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
	LogoutCallbackHandler(w http.ResponseWriter, r *http.Request)
	Init() error
}

var (
	verifier *oidc.IDTokenVerifier
)

func processAuth(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	var raw string

	auth := r.Header.Get("Authorization")
	s, err := r.Cookie("jwt")

	if err != nil && len(auth) == 0 {
		return nil, model.ErrUnauthorized
	}

	if len(auth) > 0 {
		raw = strings.Replace(auth, "Bearer ", "", 1)
	} else {
		raw = s.Value
	}

	claims := model.AuthClaims{}

	// verify claims
	idToken, err := verifier.Verify(r.Context(), raw)
	if err != nil {
		log.Error().Err(err).Msgf("could not verify jwt")

		// remove offending jwt cookie
		cookie := http.Cookie{
			Name:     "jwt",
			Value:    "",
			Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)

		return nil, err
	}

	if err := idToken.Claims(&claims); err != nil {
		// claims aren't there?
		return nil, err
	}

	// get the current user object from the database
	user := &model.User{AuthUUID: claims.Subject}
	result := model.DB.Model(user).Where(user).First(user)
	if result.Error != nil {
		return nil, fmt.Errorf("could not fetch user data for %s (%v)", claims.Subject, result.Error)
	}

	// add claims and user object to the context
	ctx := context.WithValue(r.Context(), model.AuthContextKey, claims)
	ctx = context.WithValue(ctx, util.UserContextKey, user)

	return ctx, nil
}

// Authenticated checks if there is a valid json web token in the request
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(model.AuthContextKey); v == nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Authorization header/session cookie missing"))
			return
		}

		// context is set, everything is fine
		next.ServeHTTP(w, r)
	})
}

// Authenticate checks if an authorization header or cookie is present and processes it
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := processAuth(w, r)
		if err != nil {
			if !errors.Is(err, model.ErrUnauthorized) {
				log.Error().Err(err).Msg("could not process authentication cookie/header")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, err)
			}
		}

		// everything went fine
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createUser(claims *model.AuthClaims) {
	u := model.User{
		AuthUUID: claims.Subject,
		Handle:   claims.Subject,
	}

	// set
	if claims.Nickname != "" {
		u.Handle = claims.Nickname
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
