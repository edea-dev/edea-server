package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/util"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

// Model interface defines which methods our models need to implement
type Model interface {
	MarshalZerologObject(e *zerolog.Event) // logger method to attach important information to potential log output
	GetMembers() ([]*User, error)          // can return empty list if there's just none
	GetModules() ([]*Module, error)        // can return empty list if there's just none
	Validate(u *User) error                // validate if a user is allowed to make those changes and if it makes sense
}

// DB is the global instance of the database connection
var DB *gorm.DB

// CreateTables initially creates the tables in the database
func CreateTables() {
	err := DB.AutoMigrate(&User{}, &Profile{}, &Module{}, &Repository{}, &BenchModule{}, &Category{}, &Bench{})
	if err != nil {
		log.Fatal().Err(err).Msgf("could not run automigrations")
	}

}

func isAuthorized(ctx context.Context, userID uuid.UUID, o zerolog.LogObjectMarshaler) error {
	u := ctx.Value(util.UserContextKey).(*User)

	// log if it's done by an admin
	if u.IsAdmin {
		log.Info().EmbedObject(o).Str("admin_auth_uuid", u.AuthUUID).Msg("information changed by admin")
	} else if userID != u.ID {
		log.Error().Msgf("user %s tried to change a model of %s", u.ID, userID)
		return ErrUnauthorized
	}

	return nil
}
