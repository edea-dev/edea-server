package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Bench contains a number of modules with their configuration
type Bench struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4()" schema:"id"`
	UserID      uuid.UUID     `gorm:"type:uuid" schema:"-"`
	ShortCode   string        `schema:"short_code"`
	User        User          `schema:"-"`
	Active      bool          `schema:"active"` // i.e. only show current active bench
	Public      bool          `schema:"public"`
	CreatedAt   time.Time     `schema:"-"`
	UpdatedAt   time.Time     `schema:"-"`
	Modules     []BenchModule `schema:"-"`
	Name        string        `schema:"name,required"`
	Description string        `schema:"description"`
}

// MarshalZerologObject provides the object representation for logging
func (b *Bench) MarshalZerologObject(e *zerolog.Event) {
	e.Str("bench_sc", b.ShortCode)
}

// BeforeUpdate checks if the current user is allowed to do that
func (b *Bench) BeforeUpdate(tx *gorm.DB) (err error) {
	// ctx := tx.Statement.Context

	return nil // isAuthorized(ctx, b.UserID, b)
}
