package model

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Category model
type Category struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name        string
	Description string
}

// MarshalZerologObject provides the object representation for logging
func (c *Category) MarshalZerologObject(e *zerolog.Event) {
	e.Str("category_uuid", c.ID.String())
}
