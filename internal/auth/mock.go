package auth

// SPDX-License-Identifier: EUPL-1.2

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/edea-dev/edea-server/internal/config"
	"gitlab.com/edea-dev/edea-server/internal/view"
	"go.uber.org/zap"
	jose "gopkg.in/square/go-jose.v2"
)

/*
How this works:

	Endpoints:
	/.well-known/openid-configuration	| points to the endpoints below
	/auth								| user authenticates with username and password here, returns a bearer token
	/token								| exchanges bearer token from /auth for access, id and refresh tokens
	/keys								| returns the public part of our JWKS
	/userinfo							| returns an ID token of the user

A good guide that explains the whole flow: https://connect2id.com/learn/openid-connect
*/

var (
	keySet     *jose.JSONWebKeySet
	mockSigner jose.Signer
	// user info map
	mockUsers = map[string]mockUser{
		"alice": {
			Subject:       "alice",
			Profile:       "alice",
			Email:         "alice@acme.co",
			EmailVerified: true,
			IsAdmin:       false,
			Password:      "alicealice",
		},
		"bob": {
			Subject:       "bob",
			Profile:       "bob",
			Email:         "bob@acme.co",
			EmailVerified: true,
			IsAdmin:       false,
			Password:      "bob",
		},
		"admin": {
			Subject:       "admin",
			Profile:       "admin",
			Email:         "admin@acme.co",
			EmailVerified: true,
			IsAdmin:       true,
			Password:      "12345",
		},
	}
	CallbackURL string
	Endpoint    string // where our mock OIDC server resides

	// the amount of time a client has to exchange a code for a token
	grantLifetime = time.Minute

	accessTokenLifetime = time.Hour
	idTokenLifetime     = time.Hour * 24 * 7

	// authorization grant codes
	codes = make(map[string]grant)
)

type grant struct {
	sub string
	exp time.Time
}

type mockUser struct {
	Subject       string `json:"sub"`
	Profile       string `json:"profile"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	IsAdmin       bool   `json:"is_admin"`
	Password      string `json:"password"`
}

type accessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	IDToken      string `json:"id_token,omitempty"`
}

type oidcToken struct {
	Subject       string `json:"sub"`
	Issuer        string `json:"iss"`
	Audience      string `json:"aud"`
	Nonce         string `json:"nonce,omitempty"`
	AuthTime      int    `json:"auth_time,omitempty"`
	ACR           string `json:"acr,omitempty"`
	Profile       string `json:"profile,omitempty"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	IssuedAt      int64  `json:"iat"`
	Expires       int64  `json:"exp,omitempty"`
}

// InitMockAuth initialises a keyset and provides a new mock authenticator
func InitMockAuth() error {
	var priv jose.JSONWebKey

	if keySet == nil {
		// load existing keyset if it exists
		info, err := os.Stat("mock-jwks.json")
		if !os.IsNotExist(err) && !info.IsDir() {
			s := struct {
				Priv   jose.JSONWebKey
				KeySet jose.JSONWebKeySet
			}{}
			f, err := os.Open("mock-jwks.json")
			if err != nil {
				zap.L().Fatal("could not read jwks from disk", zap.Error(err))
			}
			defer f.Close()
			dec := json.NewDecoder(f)

			if err := dec.Decode(&s); err != nil {
				zap.L().Fatal("could not decode jwks from json", zap.Error(err))
			}

			priv = s.Priv
			keySet = new(jose.JSONWebKeySet)
			*keySet = s.KeySet
		} else {
			// or generate a new one if it doesn't
			privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				zap.L().Panic("could not generate private key", zap.Error(err))
			}

			priv = jose.JSONWebKey{Key: privKey, Algorithm: "ES256", Use: "sig"}

			// Generate a canonical kid based on RFC 7638
			thumb, err := priv.Thumbprint(crypto.SHA256)
			if err != nil {
				zap.L().Panic("unable to compute thumbprint", zap.Error(err))
			}
			priv.KeyID = base64.URLEncoding.EncodeToString(thumb)

			// build our key set from the private key
			keySet = &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{priv.Public()}}

			// write the keyset to disk so we can load it later on
			f, err := os.Create("mock-jwks.json")
			if err != nil {
				zap.L().Fatal("could not save jwks to disk", zap.Error(err))
			}
			defer f.Close()
			enc := json.NewEncoder(f)
			enc.SetIndent("", "\t")

			s := struct {
				Priv   jose.JSONWebKey
				KeySet jose.JSONWebKeySet
			}{priv, *keySet}

			if err := enc.Encode(s); err != nil {
				zap.L().Fatal("could not encode jwks to json", zap.Error(err))
			}
		}

		// build a signer from our private key
		opt := (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", priv.KeyID)
		mockSigner, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: priv.Key}, opt)
		if err != nil {
			zap.L().Panic("could not create new signer", zap.Error(err))
		}
	}

	return nil
}

// LoginFormHandler provides a simple local login form for test purposes
func LoginFormHandler(c *gin.Context) {
	zap.L().Debug("authorisation ep",
		zap.String("response_type", c.Query("response_type")),
		zap.String("scope", c.Query("scope")),
		zap.String("client_id", c.Query("client_id")),
	)

	// display different login page based on client id here

	m := map[string]interface{}{
		"State":       c.Query("state"),
		"RedirectURI": c.Query("redirect_uri"),
	}

	view.RenderTemplate(c, "mock_login.tmpl", "EDeA - Login", m)
}

// LoginPostHandler processes the login request
func LoginPostHandler(c *gin.Context) {
	state := c.PostForm("state")
	user := c.PostForm("user")
	pass := c.PostForm("password")
	redirectURI := c.PostForm("redirect_uri")

	// do a basic auth check, this is the place to add a user database
	uo, ok := mockUsers[user]
	if ok && uo.Password == pass {
		if uo.Profile != user {
			zap.S().Panicf("invalid user/password combination for %s", user)
		}
	} else {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// TODO: make redirect url a []string
	if redirectURI != cfg.RedirectURL {
		zap.L().Error("got invalid redirect_uri from client", zap.String("redirect_uri", redirectURI))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	u, err := url.Parse(redirectURI)
	if err != nil {
		zap.L().Panic("could not parse callback url for mock auth", zap.Error(err))
	}

	ref := u.Query()
	ref.Set("state", state)

	// generate our authorization grant code
	code, err := uuid.NewRandom()
	if err != nil {
		zap.L().Panic("could not generate uuid code", zap.Error(err))
	}

	// store the authorization grant, temporarily
	codes[code.String()] = grant{user, time.Now().Add(grantLifetime)}

	// TODO: are we POST-ing this? can we add parameters to a POST redirect?
	ref.Set("code", code.String())
	u.RawQuery = ref.Encode()

	c.Redirect(http.StatusTemporaryRedirect, u.String())
}

// WellKnown provides the URLs of our endpoints, should be accessible at "/.well-known/openid-configuration"
func WellKnown(c *gin.Context) {
	c.String(http.StatusOK, `{
		"issuer": "%[1]s",
		"authorization_endpoint": "%[1]s/auth",
		"token_endpoint": "%[1]s/token",
		"jwks_uri": "%[1]s/keys",
		"userinfo_endpoint": "%[1]s/userinfo",
		"id_token_signing_alg_values_supported": ["ES256"]
	}`, cfg.ProviderURL)
}

// Keys endpoint provides our JSON Web Key Set (should be at /keys)
func Keys(c *gin.Context) {
	c.JSONP(http.StatusOK, keySet)
}

func generateToken(user mockUser, expires time.Duration, info bool) (string, error) {
	now := time.Now()
	exp := now.Add(expires)

	tok := oidcToken{
		Subject:  user.Subject,
		Issuer:   config.Cfg.Auth.OIDC.ProviderURL,
		Audience: config.Cfg.Auth.OIDC.ClientID,
		IssuedAt: now.Unix(),
		Expires:  exp.Unix(),
	}

	if info {
		tok.Email = user.Email
		tok.Profile = user.Profile
		tok.EmailVerified = user.EmailVerified
	}

	b, _ := json.Marshal(&tok)

	sig, _ := mockSigner.Sign(b)
	return sig.CompactSerialize()
}

// Userinfo endpoint provides the claims for a logged in user given a bearer token
// returns an id_token
func Userinfo(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	raw := strings.Replace(auth, "Bearer ", "", 1)

	// here would be the place to verify the bearer token against the issued ones
	// instead of using just static tokens which double as passwords

	zap.L().Debug("userinfo bearer", zap.String("token", raw))

	user, ok := mockUsers[raw]
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	s, err := generateToken(user, time.Hour, true)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "application/jwt")
	c.String(http.StatusOK, s)
}

// Token exchanges a "code" against a token which contains the id_token of the requested user specified in "code"
func Token(c *gin.Context) {
	code := c.PostForm("code")
	grantType := c.PostForm("grant_type")

	zap.L().Debug("token ep", zap.String("grant_type", grantType))

	// TODO: check which tokens are being requested

	g, ok := codes[code]
	if !ok {
		zap.L().Debug("token exchange unauthorized")
		c.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	// codes are single-use only
	delete(codes, code)

	if g.exp.Before(time.Now()) {
		c.String(http.StatusUnauthorized, "code expired")
		return
	}

	auth, err := generateToken(mockUsers[g.sub], accessTokenLifetime, false)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	id, err := generateToken(mockUsers[g.sub], idTokenLifetime, true)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tok := accessToken{auth, "Bearer", "", int64(accessTokenLifetime / time.Millisecond), id}

	// return token
	c.JSONP(http.StatusOK, tok)
}

// LogoutEndpoint handles logging out the user, e.g. this should invalidate
// the token auth-side so that if it is presented to us again we know that it
// has been invalidated
func LogoutEndpoint(c *gin.Context) {
	// TODO: blacklist jti of the access token here for the remaining duration
	u := c.PostForm("post_logout_redirect_uri")
	zap.S().Infof("got to logout, redirecting to: %s", u)
	c.Redirect(http.StatusTemporaryRedirect, u)
}
