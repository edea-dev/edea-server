package repo

import (
	"os"
	"testing"
	"time"

	epg "github.com/fergusstrange/embedded-postgres"
	"github.com/go-pg/pg/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/model"
)

var cfg config.Config

func TestMain(m *testing.M) {
	err := envconfig.Process("", &cfg)
	if err != nil {
		os.Exit(1)
	}

	postgres := epg.NewDatabase(epg.DefaultConfig().
		Username("postgres").
		Password("postgres").
		Database("edea").
		Version("12.3.0").
		RuntimePath("./tmp").
		Port(9877).
		StartTimeout(45 * time.Second).
		Locale("en_US.UTF-8"))
	if err := postgres.Start(); err != nil {
		log.Error().Err(err).Msg("could not start embedded postgres")
		os.Exit(1)
	}

	// start connection pool
	model.DB = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Database: "edea",
		Addr:     "127.0.0.1:9877",
	})

	if _, err := model.DB.Exec("CREATE EXTENSION pgcrypto;"); err != nil {
		log.Error().Err(err).Msg("failed to create pgcrypto extension")
		os.Exit(1)
	}

	model.CreateTables()

	cache = &RepoCache{Base: "./tmp/git"}

	code := m.Run()

	postgres.Stop()

	os.Exit(code)
}
