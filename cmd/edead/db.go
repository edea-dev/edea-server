package main

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func db() error {
	var err error

	// start connection pool
	dsn := config.Cfg.DSN
	model.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if result := model.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); result.Error != nil {
		log.Error().Err(err).Msg("failed to create uuid-ossp extension, please create it manually")
		// return err
	}

	model.CreateTables()

	return nil
}
