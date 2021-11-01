package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TODO: handle the case where we might want to link multiple users
// to a single identity. e.g. to migrate accounts or to use OIDC
// and WebAuth simultanously.

// User mapping from IDs to the authentication provider data
type User struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	AuthUUID  string    `gorm:"unique"` // unique id from authentication provider
	Handle    string    `gorm:"unique"` // user handle as it will be used in the url
	CreatedAt time.Time
	UpdatedAt time.Time
	IsAdmin   bool `gorm:"default:false"`
}

// MarshalZerologObject provides the object representation for logging
func (u *User) MarshalZerologObject(e *zerolog.Event) {
	e.Str("uuid", u.ID.String()).
		Str("auth_uuid", u.AuthUUID).
		Str("handle", u.Handle).
		Time("created", u.CreatedAt)
}

// BeforeUpdate checks if the current user is allowed to do that
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context

	return isAuthorized(ctx, u.ID, u)
}

// UserExists returns true if a user with the given auth uuid exists
func UserExists(authUUID string) bool {
	u := User{AuthUUID: authUUID}

	result := DB.Model(&u).Where(&u).First(&u)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false
		}
		log.Panic().Err(result.Error).Msgf("could not get user")
	}

	return true
}
