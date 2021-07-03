package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gitlab.com/edea-dev/edead/auth"
	mw "gitlab.com/edea-dev/edead/middleware"
)

func middleware(r *mux.Router) {
	r.Use(
		mw.RecoveryHandler(mw.PrintRecoveryStack(true), mw.PrintRoutes(true, r)),
		handlers.ProxyHeaders,
		auth.Authenticate,
	)
}
