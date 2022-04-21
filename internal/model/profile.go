package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Profile data for Users
type Profile struct {
	ID          string    `gorm:"type:uuid;primarykey;default:uuid_generate_v4()" form:"id" binding:"required,uuid"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	User        User
	DisplayName string `form:"display_name" binding:"required"`
	Location    string `form:"location"`
	Biography   string `form:"biography"`
	Avatar      string `form:"avatar"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
}

// BeforeUpdate checks if the current user is allowed to do that
func (p *Profile) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context.(*gin.Context)

	return isAuthorized(ctx, p.UserID, p)
}
