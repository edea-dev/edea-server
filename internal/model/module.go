package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Module model
type Module struct {
	ID          uuid.UUID `gorm:"type:uuid;primarykey;default:uuid_generate_v4()"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	ShortCode   string    `form:"short_code"`
	User        User
	Private     bool   `gorm:"default:false" form:"private"`
	RepoURL     string `gorm:"uniqueIndex:idx_repo_sub" form:"repourl,required"`
	Name        string `form:"name,required"`
	Sub         string `gorm:"uniqueIndex:idx_repo_sub" form:"sub"`
	Description string `form:"description"`
	CategoryID  string `gorm:"type:uuid" form:"category"`
	Category    Category
	Metadata    datatypes.JSONMap

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
}

// BeforeUpdate checks if the current user is allowed to do that
func (m *Module) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context.(*gin.Context)

	var tm Module
	result := tx.First(&tm, m.ID)
	if result.Error != nil {
		return err
	}

	return isAuthorized(ctx, tm.UserID)
}
