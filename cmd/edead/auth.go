package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/internal/auth"
	"gitlab.com/edea-dev/edead/internal/config"
)

func initAuth() {
	a := config.Cfg.Auth.OIDC
	k := config.Cfg.Auth.Kratos

	provider := &auth.OIDC{
		ClientID:      a.ClientID,
		ClientSecret:  a.ClientSecret,
		RedirectURL:   a.RedirectURL,
		ProviderURL:   a.ProviderURL,
		LogoutURL:     a.LogoutURL,
		PostLogoutURL: a.PostLogoutURL,
		OIDCConfig: &oidc.Config{
			ClientID: a.ClientID,
		},
	}

	// TODO: implement the full set of config options from auth.OIDC

	if provider.PostLogoutRedirectURIField == "" {
		provider.PostLogoutRedirectURIField = "post_logout_redirect_uri"
	}
	if provider.PostLoginRedirectURIField == "" {
		provider.PostLoginRedirectURIField = "post_login_redirect_uri"
	}

	if config.Cfg.Auth.UseMock {
		log.Warn().Msg("using mock authentication provider")
		auth.InitMockAuth()
	}

	if err := auth.Init(provider); err != nil {
		log.Error().Err(err).Msg("could not create OIDC provider")
	}

	if k.Use {
		if err := auth.InitKratos(); err != nil {
			log.Error().Err(err).Msg("could not create auth provider for Kratos")
		}
	}
}
