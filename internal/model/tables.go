package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func CreateCategories() {
	var categories = []Category{
		{Name: "Uncategorized", Description: "Not yet categorized"},
		{Name: "Power", Description: "Power electronics such as LDOs or DC/DC modules"},
		{Name: "MCU", Description: "Microcontroller modules"},
		{Name: "Test", Description: "Test modules - do not use"},
		{Name: "Connector", Description: "Connector modules"},
	}
	result := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&categories)
	if result.Error != nil {
		zap.L().Fatal("could not create categories", zap.Error(result.Error))
	}
}

func isAuthorized(c *gin.Context, userID uuid.UUID) error {
	u := c.Keys["user"].(*User)

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
