package main

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func db() error {
	var err error

	// start connection pool
	dsn := "host=192.168.0.2 user=edea password=edea dbname=edea port=5432 sslmode=disable"
	model.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if result := model.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); result.Error != nil {
		log.Error().Err(err).Msg("failed to create pgcrypto extension")
		// return err
	}

	model.CreateTables()

	return nil
}
