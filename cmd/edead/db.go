package main

import (
	"context"
	"time"

	epg "github.com/fergusstrange/embedded-postgres"
	"github.com/go-pg/pg/v10"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
)

// DebugHook is a query hook that logs the query and the error if there are any.
// It can be installed with:
//
//   db.AddQueryHook(pgext.DebugHook{})
type DebugHook struct{}

var _ pg.QueryHook = (*DebugHook)(nil)

func (DebugHook) BeforeQuery(ctx context.Context, evt *pg.QueryEvent) (context.Context, error) {
	q, err := evt.FormattedQuery()
	if err != nil {
		return nil, err
	}

	if evt.Err != nil {
		log.Debug().Msgf("Error %s executing query:\n%s\n", evt.Err, q)
	} else {
		log.Debug().Msgf("%s", q)
	}

	return ctx, nil
}

func (DebugHook) AfterQuery(context.Context, *pg.QueryEvent) error {
	return nil
}

func db() *epg.EmbeddedPostgres {
	postgres := epg.NewDatabase(epg.DefaultConfig().
		Username("postgres").
		Password("postgres").
		Database("edea").
		Version("12.3.0").
		RuntimePath("./tmp").
		Port(9876).
		StartTimeout(45 * time.Second).
		Locale("en_US.UTF-8"))
	if err := postgres.Start(); err != nil {
		log.Error().Err(err).Msg("could not start embedded postgres")
		return nil
	}

	// start connection pool
	model.DB = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Database: "edea",
		Addr:     "127.0.0.1:9876",
	})

	if _, err := model.DB.Exec("CREATE EXTENSION pgcrypto;"); err != nil {
		log.Error().Err(err).Msg("failed to create pgcrypto extension")
		return nil
	}

	model.DB.AddQueryHook(DebugHook{})

	model.CreateTables()

	return postgres
}
