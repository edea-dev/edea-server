package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func (u *User) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("uuid", u.ID.String())
	enc.AddString("auth_uuid", u.AuthUUID)
	enc.AddString("handle", u.Handle)
	enc.AddTime("created", u.CreatedAt)

	return nil
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
		zap.L().Panic("could not get user", zap.Error(result.Error))
	}

	return true
}
