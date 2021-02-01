package repo

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var cfg config.Config

func TestMain(m *testing.M) {
	err := envconfig.Process("", &cfg)
	if err != nil {
		os.Exit(1)
	}

	// start connection pool
	dsn := "host=192.168.0.2 user=edea password=edea dbname=edea port=5432 sslmode=disable"
	model.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if result := model.DB.Exec(`CREATE EXTENSION "uuid-ossp";`); result.Error != nil {
		log.Error().Err(err).Msg("failed to create uuid-ossp extension")
		os.Exit(1)
	}

	model.CreateTables()

	cache = &RepoCache{Base: "./tmp/git"}

	code := m.Run()

	os.Exit(code)
}
