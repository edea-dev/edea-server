package model

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	err := DB.AutoMigrate(&User{}, &Profile{}, &Module{}, &Repository{}, &BenchModule{})
	if err != nil {
		log.Fatal().Err(err).Msgf("could not run automigrations")
	}
}

func isAuthorized(ctx context.Context, userID uuid.UUID, o zerolog.LogObjectMarshaler) error {
	claims := ctx.Value(AuthContextKey).(AuthClaims)
	u := &User{AuthUUID: claims.Subject}

	result := DB.Model(u).Where(u).Find(u)
	if result.Error != nil {
		log.Warn().Err(result.Error).Msg("error while querying user by auth_uuid")
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrUnauthorized
		}
		return result.Error
	}

	// log if it's done by an admin
	if u.IsAdmin {
		log.Info().EmbedObject(o).Str("admin_auth_uuid", claims.Subject).Msg("information changed by admin")
	} else if userID != u.ID {
		log.Error().Msgf("user %s tried to change a model of %s", u.ID, userID)
		return ErrUnauthorized
	}

	return nil
}
