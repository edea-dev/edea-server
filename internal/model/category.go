package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category model
type Category struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name        string    `gorm:"unique"`
	Description string
}
