package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Module model
type Module struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	ShortCode   string    `form:"short_code"`
	User        User
	Private     bool      `gorm:"default:false" form:"private"`
	RepoURL     string    `form:"repourl,required"`
	Name        string    `form:"name,required"`
	Sub         string    `form:"sub"`
	Description string    `form:"description"`
	CategoryID  uuid.UUID `gorm:"type:uuid" form:"category"`
	Category    Category
	Metadata    datatypes.JSON
}

// BeforeUpdate checks if the current user is allowed to do that
func (m *Module) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context.(*gin.Context)

	return isAuthorized(ctx, m.UserID, m)
}
