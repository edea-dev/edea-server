package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"gitlab.com/edea-dev/edead/internal/config"
	"gitlab.com/edea-dev/edead/internal/model"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

func db() error {
	var err error

	// start connection pool
	dsn := config.Cfg.DSN

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		zap.L().Fatal("could not parse db config", zap.String("dsn", dsn))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	dbConn := stdlib.OpenDB(*cfg)
	if err != nil {
		zap.L().Fatal("could not open db connection", zap.Error(err))
	}

	for {
		if err := dbConn.PingContext(ctx); err != nil {
			zap.S().Info("database not yet ready", zap.Error(err))
			select {
			case <-ctx.Done():
				zap.L().Fatal("timed out waiting for database")
			case <-time.After(time.Second):
			}
		} else {
			cancel()
			break
		}
	}

	gl := zapgorm2.New(zap.L())
	gl.SetAsDefault()
	gl.IgnoreRecordNotFoundError = true
	gl.LogLevel = logger.Info

	model.DB, err = gorm.Open(postgres.New(postgres.Config{Conn: dbConn}), &gorm.Config{Logger: gl})

	if result := model.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); result.Error != nil {
		zap.L().Error("failed to create uuid-ossp extension, please create it manually", zap.Error(err))
		// return err
	}

	model.CreateTables()
	model.CreateCategories()

	return nil
}
