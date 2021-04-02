package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/repo"
)

func main() {
	var wait time.Duration
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config.ReadConfig()

	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	routes(r)
	logger(r)
	middleware(r)

	// start embedded postgres DB
	err := db()
	if err != nil {
		os.Exit(1)
	}

	// log.Info().Interface("config", config.Cfg)
	repo.InitCache(config.Cfg.Cache.Repo.Base)

	addr := fmt.Sprintf("%s:%s", config.Cfg.Server.Host, config.Cfg.Server.Port)

	srv := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("Listening on: http://%s", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Print(err)
		}
	}()

	// start out auth provider after the http server is running
	// it needs the mock auth paths already available in case it is used
	initAuth()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("could not shut down: %v", err)
		os.Exit(1)
	}

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Print("shutting down")
	os.Exit(0)
}
