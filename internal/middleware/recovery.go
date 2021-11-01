package middleware

// SPDX-License-Identifier: EUPL-1.2

import (
	"io"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"text/template"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// RecoveryHandlerLogger is an interface used by the recovering handler to print logs.
type RecoveryHandlerLogger interface {
	Println(io.Writer, ...interface{})
}

type recoveryHandler struct {
	handler     http.Handler
	logger      RecoveryHandlerLogger
	printStack  bool
	printRoutes bool
	router      *mux.Router
}

type Route struct {
	Path          string
	PathRegexp    string
	Queries       []string
	QueriesRegexp []string
	Methods       []string
	Func          string
}

// RecoveryOption provides a functional approach to define
// configuration for a handler; such as setting the logging
// whether or not to print strack traces on panic.
type RecoveryOption func(http.Handler)

func parseRecoveryOptions(h http.Handler, opts ...RecoveryOption) http.Handler {
	for _, option := range opts {
		option(h)
	}

	return h
}

// RecoveryHandler is HTTP middleware that recovers from a panic,
// logs the panic, writes http.StatusInternalServerError, and
// continues to the next handler.
//
// Example:
//
//  r := mux.NewRouter()
//  r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//  	panic("Unexpected error!")
//  })
//
//  http.ListenAndServe(":1123", handlers.RecoveryHandler()(r))
func RecoveryHandler(opts ...RecoveryOption) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		r := &recoveryHandler{handler: h}
		return parseRecoveryOptions(r, opts...)
	}
}

// RecoveryLogger is a functional option to override
// the default logger
func RecoveryLogger(logger RecoveryHandlerLogger) RecoveryOption {
	return func(h http.Handler) {
		r := h.(*recoveryHandler)
		r.logger = logger
	}
}

// PrintRecoveryStack is a functional option to enable
// or disable printing stack traces on panic.
func PrintRecoveryStack(print bool) RecoveryOption {
	return func(h http.Handler) {
		r := h.(*recoveryHandler)
		r.printStack = print
	}
}

// PrintRoutes displays all the available routes when a panic occurs
func PrintRoutes(print bool, router *mux.Router) RecoveryOption {
	return func(h http.Handler) {
		r := h.(*recoveryHandler)
		r.printRoutes = print
		r.router = router
	}
}

func routes(router *mux.Router) (m map[string]*Route) {
	m = make(map[string]*Route)
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		rr := &Route{}
		rr.Path, _ = route.GetPathTemplate()
		rr.PathRegexp, _ = route.GetPathRegexp()
		rr.Queries, _ = route.GetQueriesTemplates()
		rr.QueriesRegexp, _ = route.GetQueriesRegexp()
		rr.Methods, _ = route.GetMethods()
		rr.Func = runtime.FuncForPC(reflect.ValueOf(route.GetHandler()).Pointer()).Name()
		m[rr.Path] = rr
		return nil
	})
	return
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			r.ParseForm()
			m := map[string]interface{}{
				"context":    r.Context(),
				"stacktrace": string(debug.Stack()),
				"error":      err,
				"route":      mux.CurrentRoute(r),
				"vars":       mux.Vars(r),
				"form":       r.Form,
				"headers":    r.Header,
			}

			if h.router != nil && h.printRoutes {
				m["routes"] = routes(h.router)
			}

			tmpl, err := template.New("error tmpl").Parse(devErrorTmpl)
			if err != nil {
				zap.L().Fatal(`error while parsing panic template 🤦‍♀️🤦🤦🤦🤦🤦`, zap.Error(err))
			}
			if err := tmpl.Execute(w, m); err != nil {
				zap.L().Fatal(`error while rendering panic template 🤦‍♀️🤦🤦🤦🤦🤦`, zap.Error(err))
			}

			zap.L().Panic("recovery handler", zap.Error(err)) // TODO: add back r.Context()
		}
	}()

	h.handler.ServeHTTP(w, r)
}
