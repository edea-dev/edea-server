package auth

// SPDX-License-Identifier: EUPL-1.2

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/view"
	"gopkg.in/square/go-jose.v2"
)

// MockAuth mock authentication handler, logs in user as "acme"
var (
	keySet      *jose.JSONWebKeySet
	mockSigner  jose.Signer
	mockUsers   map[string]mockUser // user info map
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
				log.Fatal().Err(err).Msg("could not read jwks from disk")
			}
			defer f.Close()
			dec := json.NewDecoder(f)

			if err := dec.Decode(&s); err != nil {
				log.Fatal().Err(err).Msg("could not decode jwks from json")
			}

			priv = s.Priv
			keySet = new(jose.JSONWebKeySet)
			*keySet = s.KeySet
		} else {
			// or generate a new one if it doesn't
			privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				log.Panic().Err(err).Msg("could not generate private key")
			}

			priv = jose.JSONWebKey{Key: privKey, Algorithm: "ES256", Use: "sig"}

			// Generate a canonical kid based on RFC 7638
			thumb, err := priv.Thumbprint(crypto.SHA256)
			if err != nil {
				log.Panic().Err(err).Msg("unable to compute thumbprint")
			}
			priv.KeyID = base64.URLEncoding.EncodeToString(thumb)

			// build our key set from the private key
			keySet = &jose.JSONWebKeySet{Keys: []jose.JSONWebKey{priv.Public()}}

			// write the keyset to disk so we can load it later on
			f, err := os.Create("mock-jwks.json")
			if err != nil {
				log.Fatal().Err(err).Msg("could not save jwks to disk")
			}
			defer f.Close()
			enc := json.NewEncoder(f)
			enc.SetIndent("", "\t")

			s := struct {
				Priv   jose.JSONWebKey
				KeySet jose.JSONWebKeySet
			}{priv, *keySet}

			if err := enc.Encode(s); err != nil {
				log.Fatal().Err(err).Msg("could not encode jwks to json")
			}
		}

		// build a signer from our private key
		opt := (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", priv.KeyID)
		mockSigner, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: priv.Key}, opt)
		if err != nil {
			log.Panic().Err(err).Msg("could not create new signer")
		}
	}

	return nil
}

// LoginFormHandler provides a simple local login form for test purposes
func LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic().Err(err).Msg("could not parse url parameters for the login form")
	}

	m := map[string]interface{}{
		"State": r.Form.Get("state"),
	}

	// TODO: show a simple login form
	view.RenderTemplate(r.Context(), "mock_login.tmpl", "EDeA - Login", m, w)
}

// LoginPostHandler processes the login request
func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic().Err(err).Msg("could not parse login form parameters")
	}

	state := r.Form.Get("state")
	user := r.Form.Get("user")
	pass := r.Form.Get("password")

	// do a basic auth "check", this *really* is just a mock authenticator
	if u, ok := mockUsers[pass]; ok {
		if u.Profile != user {
			log.Panic().Msgf("invalid user/password combination for %s", user)
		}
	}

	u, err := url.Parse(cfg.RedirectURL)
	if err != nil {
		log.Panic().Err(err).Msg("could not parse callback url for mock auth")
	}

	ref := u.Query()
	ref.Set("state", state)
	// we could also generate a token here which is valid to get the info for the specified user
	// it would be a good starting point for a more complete OIDC server
	ref.Set("code", pass)
	u.RawQuery = ref.Encode()

	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
}

// WellKnown provides the URLs of our endpoints, should be accessible at "/.well-known/openid-configuration"
func WellKnown(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, `{
		"issuer": "%[1]s",
		"authorization_endpoint": "%[1]s/auth",
		"token_endpoint": "%[1]s/token",
		"jwks_uri": "%[1]s/keys",
		"userinfo_endpoint": "%[1]s/userinfo",
		"id_token_signing_alg_values_supported": ["ES256"]
	}`, cfg.ProviderURL)

	if err != nil {
		w.WriteHeader(500)
	}
}

// Keys endpoint provides our JSON Web Key Set (should be at /keys)
func Keys(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	if err := enc.Encode(keySet); err != nil {
		log.Error().Err(err).Msg("could not encode jwks")
		w.WriteHeader(500)
	}
}

func generateIDToken(u mockUser) (string, error) {
	b, err := json.Marshal(u)
	if err != nil {
		return "", err
	}

	sig, err := mockSigner.Sign(b)
	return sig.FullSerialize(), nil
}

// Userinfo endpoint provides the claims for a logged in user given a bearer token
func Userinfo(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	raw := strings.Replace(auth, "Bearer ", "", 1)

	// here would be the place to verify the bearer token against the issued ones
	// instead of using just static tokens which double as passwords

	s, err := generateIDToken(mockUsers[raw])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/jwt")
	_, err = io.WriteString(w, s)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Token exchanges a "code" against a token which contains the id_token of the requested user specified in "code"
func Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.FormValue("client_id")
	secret := r.FormValue("client_secret")
	code := r.FormValue("code")

	log.Debug().Msgf("form: %+v, headers: %+v", r.PostForm, r.Header)

	if cfg.ClientID == id && cfg.ClientSecret == secret {
		s, err := generateIDToken(mockUsers[code])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// return token
		tok := fmt.Sprintf(`{"access_token": "%s", "token_type": "Bearer", "refresh_token": "%s", "expires_in": 3600, "id_token": "%s"}`,
			"SlAV32hkKG",
			"8xLOxBtZp8",
			s,
		)
		if _, err := io.WriteString(w, tok); err != nil {
			log.Error().Err(err).Msg("could not send token response to client")
		}
	} else {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
}
