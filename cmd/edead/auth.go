package main

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/auth"
	"gitlab.com/edea-dev/edea/backend/config"
)

func initAuth() auth.Provider {
	var provider auth.Provider
	var err error
	a := config.Cfg.Auth.Auth0
	k := config.Cfg.Auth.Kratos
	if a.Use {
		provider, err = auth.InitAuth0(a.ClientID, a.ClientSecret, a.RedirectURL, a.ProviderURL, a.SessionSecret)
		if err != nil {
			log.Panic().Err(err).Msg("could not create auth provider for Auth0")
		}
	} else if k.Use {
		provider, err = auth.InitKratos()
		if err != nil {
			log.Panic().Err(err).Msg("could not create auth provider for Auth0")
		}
	} else {
		log.Warn().Msg("using mock authentication provider")
		provider = auth.InitMockAuth()
	}

	jwks := config.Cfg.Auth.JWKS
	if len(jwks) > 0 {
		var b []byte
		u, err := url.Parse(jwks)

		// try to parse as base64 if it's not a URL
		if err != nil {
			b, err = base64.RawURLEncoding.DecodeString(jwks)
			if err != nil {
				log.Fatal().Msg("jwks config option set but is neither an URL or base64url encoded json")
			}
		} else {
			if u.Scheme == "http" || u.Scheme == "https" {
				resp, err := http.Get(u.String())
				if err != nil {
					log.Fatal().Err(err).Msg("could not retrieve remote jwks")
				}
				defer resp.Body.Close()
				b, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal().Err(err).Msg("error while reading jwks from remote")
				}
			} else if u.Scheme == "file" {
				b, err = ioutil.ReadFile(u.Path)
				if err != nil {
					log.Fatal().Err(err).Msg("could not read jwks file")
				}
			} else {
				log.Fatal().Msg("jwks url is neither http(s) nor file")
			}
		}

		if err := auth.InitJWKS(b); err != nil {
			log.Fatal().Err(err).Msg("could not parse jwks string")
		}
	}

	return provider
}
