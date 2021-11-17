package auth

// SPDX-License-Identifier: EUPL-1.2
//
// Generic OIDC auth handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/model"
	"go.uber.org/zap"
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
		zap.L().Error("failed to get provider", zap.Error(err))
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
func CallbackHandler(c *gin.Context) {
	state, err := c.Cookie("state")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if c.Query("state") != state {
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid state parameter"))
		return
	}

	token, err := auth.Config.Exchange(c, c.Query("code"))
	if err != nil {
		zap.L().Error("no token found", zap.Error(err))
		c.Status(http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.AbortWithError(http.StatusInternalServerError, errors.New("no id_token field in oauth2 token"))
		return
	}

	tok, err := verifier.Verify(c, rawIDToken)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to verify id token: %w", err))
		return
	}

	// check if it's a new user and create them if necessary
	if !model.UserExists(tok.Subject) {
		zap.S().Debugf("user %s does not exist yet", tok.Subject)
		claims := &model.AuthClaims{}
		if err := tok.Claims(claims); err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to parse claims: %w\n%+v", err, *tok))
			return
		}

		createUser(claims)
	} else {
		zap.S().Debugf("user %s already exists", tok.Subject)
	}

	// add the jwt as session cookie
	c.SetCookie("jwt", rawIDToken, 3600, "/", "", true, false)

	// do a meta-refresh redirect so that we lose the cross-origin flag
	var html = `<html>
	<head>
	<meta http-equiv="refresh" content="0;URL='/'"/>
	</head>
	<body><p>Moved to <a href="/">Home</a>. This is just a redirect for OpenID Connect.</p></body>
	</html>
	`
	c.Header("content-type", "text/html")
	c.String(http.StatusOK, html)
}

// LoginHandler http handler
func LoginHandler(c *gin.Context) {
	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// store the nonce as a temporary cookie and allow for some time to complete the login flow
	http.SetCookie(c.Writer, &http.Cookie{Name: "state", Value: state, MaxAge: 3600})

	c.Redirect(http.StatusTemporaryRedirect, auth.Config.AuthCodeURL(state))
}

// LogoutHandler http handler
func LogoutHandler(c *gin.Context) {
	logoutURL, err := url.Parse(cfg.LogoutURL)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	parameters := url.Values{}

	parameters.Add(cfg.PostLogoutRedirectURIField, cfg.PostLogoutURL)

	if cfg.LogoutClientID {
		parameters.Add("client_id", cfg.ClientID)
	}

	if cfg.LogoutIDTokenHint {
		val, err := c.Cookie("jwt")
		if err != nil {
			zap.L().Panic("no session cookie present", zap.Error(err))
		}
		parameters.Add("id_token_hint", val)
	}

	if cfg.LogoutNonce {
		// Generate random state
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		state := base64.StdEncoding.EncodeToString(b)

		// store the nonce as a temporary cookie and allow for some time to complete the login flow
		http.SetCookie(c.Writer, &http.Cookie{Name: "state", Value: state, MaxAge: 3600})
	}

	logoutURL.RawQuery = parameters.Encode()

	c.Redirect(http.StatusTemporaryRedirect, logoutURL.String())
}

// LogoutCallbackHandler verifies the CSRF token (if set) and removes the session cookie
func LogoutCallbackHandler(c *gin.Context) {
	state, err := c.Cookie("state")
	if err == nil {
		if val, _ := c.GetQuery("state"); val != state {
			//http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			zap.S().Debugf("unexpected state value from client")
		} else {
			// remove state cookie
			cookie := http.Cookie{
				Name:     "state",
				Value:    "",
				Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(c.Writer, &cookie)
		}
	}

	// remove session cookie
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, &cookie)

	// redirect to the main page after logout
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
