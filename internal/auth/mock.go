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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/view"
	"go.uber.org/zap"
	jose "gopkg.in/square/go-jose.v2"
)

var (
	keySet     *jose.JSONWebKeySet
	mockSigner jose.Signer
	// user info map
	mockUsers = map[string]mockUser{
		"alicealice": {
			Subject:       "alice",
			Profile:       "alice",
			Email:         "alice@acme.co",
			EmailVerified: true,
			IsAdmin:       false,
		},
		"bob": {
			Subject:       "bob",
			Profile:       "bob",
			Email:         "bob@acme.co",
			EmailVerified: true,
			IsAdmin:       false,
		},
		"12345": {
			Subject:       "admin",
			Profile:       "admin",
			Email:         "admin@acme.co",
			EmailVerified: true,
			IsAdmin:       true,
		},
	}
	CallbackURL string
	Endpoint    string // where our mock OIDC server resides
)

type mockUser struct {
	Subject       string `json:"sub"`
	Profile       string `json:"profile"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	IsAdmin       bool   `json:"is_admin"`
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
	m := map[string]interface{}{
		"State": c.PostForm("state"),
	}

	// TODO: show a simple login form
	view.RenderTemplate(c, "mock_login.tmpl", "EDeA - Login", m)
}

// LoginPostHandler processes the login request
func LoginPostHandler(c *gin.Context) {
	state := c.PostForm("state")
	user := c.PostForm("user")
	pass := c.PostForm("password")

	// do a basic auth "check", this *really* is just a mock authenticator
	if u, ok := mockUsers[pass]; ok {
		if u.Profile != user {
			zap.S().Panicf("invalid user/password combination for %s", user)
		}
	}

	u, err := url.Parse(cfg.RedirectURL)
	if err != nil {
		zap.L().Panic("could not parse callback url for mock auth", zap.Error(err))
	}

	ref := u.Query()
	ref.Set("state", state)
	// we could also generate a token here which is valid to get the info for the specified user
	// it would be a good starting point for a more complete OIDC server
	ref.Set("code", pass)
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

func generateIDToken(u mockUser) (string, error) {
	b, err := json.Marshal(u)
	if err != nil {
		return "", err
	}

	sig, _ := mockSigner.Sign(b)
	return sig.FullSerialize(), nil
}

// Userinfo endpoint provides the claims for a logged in user given a bearer token
func Userinfo(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	raw := strings.Replace(auth, "Bearer ", "", 1)

	// here would be the place to verify the bearer token against the issued ones
	// instead of using just static tokens which double as passwords

	s, err := generateIDToken(mockUsers[raw])
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "application/jwt")
	c.String(http.StatusOK, s)
}

// Token exchanges a "code" against a token which contains the id_token of the requested user specified in "code"
func Token(c *gin.Context) {
	id := c.PostForm("client_id")
	secret := c.PostForm("client_secret")
	code := c.PostForm("code")

	if cfg.ClientID == id && cfg.ClientSecret == secret {
		s, err := generateIDToken(mockUsers[code])
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		// return token
		tok := fmt.Sprintf(`{"access_token": "%s", "token_type": "Bearer", "refresh_token": "%s", "expires_in": 3600, "id_token": "%s"}`,
			"SlAV32hkKG",
			"8xLOxBtZp8",
			s,
		)
		c.String(http.StatusOK, tok)
	} else {
		c.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
	}
}

// LogoutEndpoint handles logging out the user, e.g. this should invalidate
// the token auth-side so that if it is presented to us again we know that it
// has been invalidated
func LogoutEndpoint(c *gin.Context) {
	u := c.PostForm("post_logout_redirect_uri")
	zap.S().Infof("got to logout, redirecting to: %s", u)
	c.Redirect(http.StatusTemporaryRedirect, u)
}
