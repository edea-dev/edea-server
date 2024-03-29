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
	"github.com/matthewhartstonge/argon2"
	"gitlab.com/edea-dev/edea-server/internal/config"
	"gitlab.com/edea-dev/edea-server/internal/view"
	"go.uber.org/zap"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/yaml.v3"
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
	keySet *jose.JSONWebKeySet
	signer jose.Signer
	// user info map
	users       map[string]User
	CallbackURL string
	Endpoint    string // where our OIDC server resides

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

type User struct {
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

type wellKnown struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	JwksURI                          string   `json:"jwks_uri"`
	UserinfoEndpoint                 string   `json:"userinfo_endpoint"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

func stringInArray(s string, a []string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}

	return false
}

// InitOIDCServer initialises a keyset and provides a new authenticator
func InitOIDCServer() {
	var priv jose.JSONWebKey

	if keySet == nil {
		// load existing keyset if it exists
		info, err := os.Stat("uoidc-jwks.json")
		if !os.IsNotExist(err) && !info.IsDir() {
			s := struct {
				Priv   jose.JSONWebKey
				KeySet jose.JSONWebKeySet
			}{}
			f, err := os.Open("uoidc-jwks.json")
			if err != nil {
				zap.L().Fatal("could not read jwks from disk", zap.Error(err))
			}
			dec := json.NewDecoder(f)

			if err := dec.Decode(&s); err != nil {
				zap.L().Fatal("could not decode jwks from json", zap.Error(err))
			}

			_ = f.Close()

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
			f, err := os.Create("uoidc-jwks.json")
			if err != nil {
				zap.L().Fatal("could not save jwks to disk", zap.Error(err))
			}
			enc := json.NewEncoder(f)
			enc.SetIndent("", "\t")

			s := struct {
				Priv   jose.JSONWebKey
				KeySet jose.JSONWebKeySet
			}{priv, *keySet}

			if err := enc.Encode(s); err != nil {
				zap.L().Fatal("could not encode jwks to json", zap.Error(err))
			}
			_ = f.Close()
		}

		// build a signer from our private key
		opt := (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", priv.KeyID)
		signer, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: priv.Key}, opt)
		if err != nil {
			zap.L().Panic("could not create new signer", zap.Error(err))
		}

		var usersFile = "users.yml"

		// load the users from users.yml
		if config.Cfg.Auth.MiniOIDCServer.UsersFile != "" {
			usersFile = config.Cfg.Auth.MiniOIDCServer.UsersFile
		}

		info, err = os.Stat(usersFile)
		if os.IsNotExist(err) || info.IsDir() {
			zap.L().Fatal("builtin oidc auth specified but no users.yml available", zap.Error(err))
		}

		f, err := os.Open(usersFile)
		if err != nil {
			zap.L().Fatal("could not open users.yml", zap.Error(err))
		}

		dec := yaml.NewDecoder(f)
		if err := dec.Decode(&users); err != nil {
			zap.L().Fatal("could not parse users.yml", zap.Error(err))
		}
	}
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

	view.RenderTemplate(c, "builtin_login.tmpl", "EDeA - Login", m)
}

// LoginPostHandler processes the login request
func LoginPostHandler(c *gin.Context) {
	var err error
	state := c.PostForm("state")
	user := c.PostForm("user")
	pass := c.PostForm("password")
	redirectURI := c.PostForm("redirect_uri")

	// do a basic auth check, this is the place to add a user database
	uo, ok := users[user]

	if ok {
		// check password against encoded hash
		ok, err = argon2.VerifyEncoded([]byte(pass), []byte(uo.Password))
		if err != nil {
			zap.L().Panic("invalid hash in users.yml", zap.String("user", uo.Subject), zap.Error(err))
		}
		if uo.Profile != user {
			zap.S().Panicf("invalid user/password combination for %s", user)
		}
	}

	if !ok {
		// password didn't match
		c.AbortWithStatus(http.StatusForbidden)
		view.RenderTemplate(c, "403.tmpl", "Forbidden", nil)
		return
	}

	if !stringInArray(redirectURI, config.Cfg.Auth.MiniOIDCServer.RedirectURLs) {
		zap.L().Error("got invalid redirect_uri from client", zap.String("redirect_uri", redirectURI))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	u, err := url.Parse(redirectURI)
	if err != nil {
		zap.L().Panic("could not parse callback url for builtin oidc auth", zap.Error(err))
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
	c.JSONP(http.StatusOK, wellKnown{
		Issuer:                           cfg.ProviderURL,
		AuthorizationEndpoint:            cfg.ProviderURL + "/auth",
		TokenEndpoint:                    cfg.ProviderURL + "/token",
		JwksURI:                          cfg.ProviderURL + "/keys",
		UserinfoEndpoint:                 cfg.ProviderURL + "/userinfo",
		IDTokenSigningAlgValuesSupported: []string{"ES256"},
	})
}

// Keys endpoint provides our JSON Web Key Set (should be at /keys)
func Keys(c *gin.Context) {
	c.JSONP(http.StatusOK, keySet)
}

func generateToken(user User, expires time.Duration, info bool) (string, error) {
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

	sig, _ := signer.Sign(b)
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

	user, ok := users[raw]
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	s, err := generateToken(user, time.Hour, true)

	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
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
		_ = c.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	// codes are single-use only
	delete(codes, code)

	if g.exp.Before(time.Now()) {
		c.String(http.StatusUnauthorized, "code expired")
		return
	}

	auth, err := generateToken(users[g.sub], accessTokenLifetime, false)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	id, err := generateToken(users[g.sub], idTokenLifetime, true)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
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
	// loggedOutJTIs[""] = true
	u := c.PostForm("post_logout_redirect_uri")
	clientID := c.PostForm("client_id")
	idTokenHint := c.PostForm("client_id_token_hint")

	if clientID != config.Cfg.Auth.OIDC.ClientID {
		c.AbortWithStatus(http.StatusBadRequest)
		zap.L().Debug("logout unknown client id", zap.String("client_id", clientID))
		return
	}

	// check if the post logout url is valid
	if !stringInArray(u, config.Cfg.Auth.MiniOIDCServer.PostLogoutURLs) {
		c.AbortWithStatus(http.StatusBadRequest)
		zap.L().Debug("logout unknown redirect url", zap.String("logout_redirect_url", u))
		return
	}

	// verify claims
	idToken, err := verifier.Verify(c, idTokenHint)
	if err != nil {
		zap.L().Error("could not verify jwt", zap.Error(err))
	}

	zap.S().Infof("id token: %#v, token hint: %s", idToken, idTokenHint)

	zap.S().Debugf("got to logout, redirecting to: %s", u)
	c.Redirect(http.StatusTemporaryRedirect, u)
}
