package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Profile data for Users
type Profile struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	User        User
	DisplayName string `schema:"display_name,required"`
	Location    string
	Biography   string
	Avatar      string
}

// BeforeUpdate checks if the current user is allowed to do that
func (p *Profile) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context

	return isAuthorized(ctx, p.UserID, p)
}
