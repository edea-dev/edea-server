package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/square/go-jose/jwt"
	"gopkg.in/square/go-jose.v2"
)

// MockAuth mock authentication handler, logs in user as "acme"
type MockAuth struct {
}

var (
	mockSigner jose.Signer
)

// InitMockAuth initialises a keyset and provides a new mock authenticator
func InitMockAuth() *MockAuth {
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

	return &MockAuth{}
}

// LoginHandler mock function
func (a *MockAuth) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	// set up our test user claims
	cl := jwt.Claims{
		Subject:   "33695c0d-a563-4458-87a0-408854f406e3",
		Issuer:    "acme",
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		Audience:  jwt.Audience{""},
	}
	raw, err := jwt.Signed(mockSigner).Claims(cl).CompactSerialize()
	if err != nil {
		log.Panic().Err(err).Msg("could not sign test token")
	}

	// add a session cookie
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    raw,
		SameSite: http.SameSiteStrictMode,
	}
	log.Debug().Msg("trying to set cookie")
	http.SetCookie(w, &cookie)

	ref := r.Form.Get("ref")
	if len(ref) == 0 {
		ref = "/callback"
	}

	http.Redirect(w, r, ref, http.StatusTemporaryRedirect)
}

// CallbackHandler http handler
func (a *MockAuth) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Middleware checks already if the request is okay, no further action needed
	createUserIfNotExist(w, r)

	// Redirect to logged in page
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// LogoutHandler http handler
func (a *MockAuth) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
