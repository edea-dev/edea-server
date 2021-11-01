package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Module model
type Module struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	ShortCode   string    `schema:"short_code"`
	User        User
	Private     bool      `gorm:"default:false" schema:"private"`
	RepoURL     string    `schema:"repourl,required"`
	Name        string    `schema:"name,required"`
	Sub         string    `schema:"sub"`
	Description string    `schema:"description"`
	CategoryID  uuid.UUID `gorm:"type:uuid" schema:"category"`
	Category    Category
	Metadata    datatypes.JSON
}

// BeforeUpdate checks if the current user is allowed to do that
func (m *Module) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context

	return isAuthorized(ctx, m.UserID, m)
}
