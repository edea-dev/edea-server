package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"gitlab.com/edea-dev/edead/internal/auth"
	"gitlab.com/edea-dev/edead/internal/config"
	"go.uber.org/zap"
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
		zap.L().Warn("using mock authentication provider")
		auth.InitMockAuth()
	}

	if err := auth.Init(provider); err != nil {
		zap.L().Error("could not create OIDC provider", zap.Error(err))
	}

	if k.Use {
		if err := auth.InitKratos(); err != nil {
			zap.L().Error("could not create auth provider for Kratos", zap.Error(err))
		}
	}
}
