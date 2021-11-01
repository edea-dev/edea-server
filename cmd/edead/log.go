package main

// SPDX-License-Identifier: EUPL-1.2

import (
	_ "net/http/pprof"

	"github.com/gorilla/mux"
)

func logger(r *mux.Router) {
	// Install the logger handler with default output on the console

	// TODO: write a log handler for zap

	/*
		r.Use(hlog.NewHandler(log.Logger))

		// Install some provided extra handler to set some request's context fields.
		// Thanks to that handler, all our logs will come with some prepopulated fields.
		r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("")
		}))
		r.Use(hlog.RemoteAddrHandler("ip"))
		r.Use(hlog.UserAgentHandler("user_agent"))
		r.Use(hlog.RefererHandler("referer"))
		r.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

			// Here is your final handler
			h := c.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Get the logger from the request's context. You can safely assume it
				// will be always there: if the handler is removed, hlog.FromRequest
				// will return a no-op logger.
				hlog.FromRequest(r).Info().
					Str("user", "current user").
					Str("status", "ok").
					Msg("Something happened")

				// Output: {"level":"info","time":"2001-02-03T04:05:06Z","role":"my-service","host":"local-hostname","req_id":"b4g0l5t6tfid6dtrapu0","user":"current user","status":"ok","message":"Something happened"}
			}))
			http.Handle("/", h)
	*/
}
