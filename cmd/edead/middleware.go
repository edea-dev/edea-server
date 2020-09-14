package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	mw "gitlab.com/edea-dev/edea/backend/middleware"
)

func middleware(r *mux.Router) {
	r.Use(mw.RecoveryHandler(mw.PrintRecoveryStack(true), mw.PrintRoutes(true, r)), handlers.ProxyHeaders)
}
