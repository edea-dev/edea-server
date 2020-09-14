package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"net/url"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type auth0Authenticator struct {
	Provider *oidc.Provider
	Config   oauth2.Config
	Ctx      context.Context
}

// Auth0 Provider
type Auth0 struct {
	Store        *sessions.FilesystemStore
	clientID     string
	clientSecret string
	redirectURL  string
	providerURL  string
}

// InitAuth0 initialises an Auth0 provider to handle users
func InitAuth0(clientID, clientSecret, redirectURL, providerURL, sessionSecret string) (*Auth0, error) {
	a := new(Auth0)
	a.clientID = clientID
	a.clientSecret = clientSecret
	a.redirectURL = redirectURL
	a.providerURL = providerURL
	a.Store = sessions.NewFilesystemStore("", []byte(sessionSecret))
	gob.Register(map[string]interface{}{})
	return a, nil
}

func (a *Auth0) newAuthenticator() (*auth0Authenticator, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, a.providerURL)
	if err != nil {
		log.Printf("failed to get provider: %v", err)
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		RedirectURL:  a.redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &auth0Authenticator{
		Provider: provider,
		Config:   conf,
		Ctx:      ctx,
	}, nil
}

// CallbackHandler http handler
func (a *Auth0) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	authenticator, err := a.newAuthenticator()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := authenticator.Config.Exchange(context.TODO(), r.URL.Query().Get("code"))
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

	oidcConfig := &oidc.Config{
		ClientID: a.clientID,
	}

	_, err = authenticator.Provider.Verifier(oidcConfig).Verify(context.TODO(), rawIDToken)

	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// add the jwt as session cookie
	http.SetCookie(w, &http.Cookie{Name: "jwt", Value: rawIDToken, SameSite: http.SameSiteStrictMode})

	createUserIfNotExist(w, r)

	// Redirect to logged in page
	http.Redirect(w, r, "/user", http.StatusSeeOther)
}

// LoginHandler http handler
func (a *Auth0) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	authenticator, err := a.newAuthenticator()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authenticator.Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// LogoutHandler http handler
func (a *Auth0) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logoutURL, err := url.Parse(a.providerURL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logoutURL.Path += "/v2/logout"
	parameters := url.Values{}

	var scheme string
	if r.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", a.clientID)
	logoutURL.RawQuery = parameters.Encode()

	// remove session cookie
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}
