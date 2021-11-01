package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/google/uuid"
	"gitlab.com/edea-dev/edead/internal/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

// Model interface defines which methods our models need to implement
type Model interface {
	// MarshalZerologObject(e *zerolog.Event) // logger method to attach important information to potential log output
	GetMembers() ([]*User, error)   // can return empty list if there's just none
	GetModules() ([]*Module, error) // can return empty list if there's just none
	Validate(u *User) error         // validate if a user is allowed to make those changes and if it makes sense
}

// DB is the global instance of the database connection
var DB *gorm.DB

// CreateTables initially creates the tables in the database
func CreateTables() {
	err := DB.AutoMigrate(&User{}, &Profile{}, &Module{}, &Repository{}, &BenchModule{}, &Category{}, &Bench{})
	if err != nil {
		zap.L().Fatal("could not run automigrations", zap.Error(err))
	}

}

func isAuthorized(ctx context.Context, userID uuid.UUID) error {
	u := ctx.Value(util.UserContextKey).(*User)

	// log if it's done by an admin
	if u.IsAdmin {
		zap.L().Warn("information changed by admin", zap.String("admin_auth_uuid", u.AuthUUID))
	} else if userID != u.ID {
		zap.L().Error("user_a tried to change model of user_b",
			zap.String("user_a", u.ID.String()),
			zap.String("user_b", userID.String()))
		return ErrUnauthorized
	}

	return nil
}
