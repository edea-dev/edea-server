package model

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Module model
type Module struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	User        User
	Private     bool   `schema:"private,required"`
	RepoURL     string `schema:"repourl,required"`
	Name        string `schema:"name,required"`
	Description string `schema:"description"`
}

// MarshalZerologObject provides the object representation for logging
func (m *Module) MarshalZerologObject(e *zerolog.Event) {
	e.Str("module_uuid", m.ID.String())
}

// BeforeUpdate checks if the current user is allowed to do that
func (m *Module) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context

	return isAuthorized(ctx, m.User.ID, m)
}
