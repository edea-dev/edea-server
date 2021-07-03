package auth

// SPDX-License-Identifier: EUPL-1.2
//
// Generic OIDC auth handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/model"
	"golang.org/x/oauth2"
)

type authenticator struct {
	Provider *oidc.Provider
	Config   oauth2.Config
	Ctx      context.Context
}

// OIDC Provider
type OIDC struct {
	ClientID                   string
	ClientSecret               string
	RedirectURL                string
	ProviderURL                string
	LogoutURL                  string // provider logout endpoint
	PostLogoutURL              string
	LogoutIDTokenHint          bool
	LogoutNonce                bool
	LogoutClientID             bool
	PostLoginRedirectURIField  string // defaults to "post_login_redirect_uri"
	PostLogoutRedirectURIField string // defaults to "post_logout_redirect_uri"

	OIDCConfig *oidc.Config
}

var (
	auth *authenticator
	cfg  *OIDC
)

// Init the OIDC provider
func Init(a *OIDC) (err error) {
	cfg = a
	auth, err = a.newAuthenticator()
	if err == nil {
		verifier = auth.Provider.Verifier(cfg.OIDCConfig)
	}
	return err
}

func (a *OIDC) newAuthenticator() (*authenticator, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, a.ProviderURL)
	if err != nil {
		log.Printf("failed to get provider: %v", err)
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		RedirectURL:  a.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &authenticator{
		Provider: provider,
		Config:   conf,
		Ctx:      ctx,
	}, nil
}

// CallbackHandler http handler
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	token, err := auth.Config.Exchange(context.TODO(), r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("no token found: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	tok, err := verifier.Verify(context.TODO(), rawIDToken)

	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// check if it's a new user and create them if necessary
	if !model.UserExists(tok.Subject) {
		log.Debug().Msgf("user %s does not exist yet", tok.Subject)
		claims := &model.AuthClaims{}
		if err := tok.Claims(claims); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse claims: %v\n%+v", err, *tok), http.StatusInternalServerError)
			return
		}

		createUser(claims)
	} else {
		log.Debug().Msgf("user %s already exists", tok.Subject)
	}

	// add the jwt as session cookie
	http.SetCookie(w, &http.Cookie{Name: "jwt", Value: rawIDToken, SameSite: http.SameSiteStrictMode})

	// do a meta-refresh redirect so that we lose the cross-origin flag
	var html = `<html>
	<head>
	<meta http-equiv="refresh" content="0;URL='/'"/>
	</head>
	<body><p>Moved to <a href="/">Home</a>. This is just a redirect for OpenID Connect.</p></body>
	</html>
	`
	fmt.Fprintf(w, "%s", html)
}

// LoginHandler http handler
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// store the nonce as a temporary cookie and allow for some time to complete the login flow
	http.SetCookie(w, &http.Cookie{Name: "state", Value: state, MaxAge: 3600})

	http.Redirect(w, r, auth.Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// LogoutHandler http handler
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logoutURL, err := url.Parse(cfg.LogoutURL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parameters := url.Values{}

	parameters.Add(cfg.PostLogoutRedirectURIField, cfg.PostLogoutURL)

	if cfg.LogoutClientID {
		parameters.Add("client_id", cfg.ClientID)
	}

	if cfg.LogoutIDTokenHint {
		c, err := r.Cookie("jwt")
		if err != nil {
			log.Panic().Err(err).Msg("no session cookie present")
		}
		parameters.Add("id_token_hint", c.Value)
	}

	if cfg.LogoutNonce {
		// Generate random state
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		state := base64.StdEncoding.EncodeToString(b)

		// store the nonce as a temporary cookie and allow for some time to complete the login flow
		http.SetCookie(w, &http.Cookie{Name: "state", Value: state, MaxAge: 3600})
	}

	logoutURL.RawQuery = parameters.Encode()

	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}

// LogoutCallbackHandler verifies the CSRF token (if set) and removes the session cookie
func LogoutCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err == nil {
		if r.URL.Query().Get("state") != state.Value {
			//http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			log.Debug().Msgf("unexpected state value from client")
		} else {
			// remove state cookie
			cookie := http.Cookie{
				Name:     "state",
				Value:    "",
				Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, &cookie)
		}
	}

	// remove session cookie
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)

	// redirect to the main page after logout
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
