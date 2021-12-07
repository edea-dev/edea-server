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

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/config"
	"gitlab.com/edea-dev/edead/internal/middleware"
	"gitlab.com/edea-dev/edead/internal/repo"
	"gitlab.com/edea-dev/edead/internal/search"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var wait time.Duration

	zc := zap.NewDevelopmentConfig()
	zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zl, _ := zc.Build()
	defer zl.Sync()

	zap.ReplaceGlobals(zl)

	config.ReadConfig()

	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ginzap.GinzapWithConfig(zl, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/css", "/js", "/img", "/fonts", "/icons"},
	}))
	r.Use(gin.CustomRecoveryWithWriter(nil, middleware.Recovery))

	routes(r)

	// start embedded postgres DB
	err := db()
	if err != nil {
		os.Exit(1)
	}

	// zap.S().Info().Interface("config", config.Cfg)
	repo.InitCache(config.Cfg.Cache.Repo.Base)

	if err := search.Init(config.Cfg.Search.Host, config.Cfg.Search.Index, config.Cfg.Search.APIKey); err != nil {
		zap.L().Error("could not init search", zap.Error(err))
	}

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
		zap.S().Infof("Listening on: http://%s", addr)
		if err := srv.ListenAndServe(); err != nil {
			zap.L().Error("could not listen", zap.Error(err))
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
		zap.L().Error("could not shut down", zap.Error(err))
		os.Exit(1)
	}

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	zap.S().Info("shutting down")
	os.Exit(0)
}
