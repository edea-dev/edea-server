package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/google/uuid"
)

// Category model
type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primarykey;default:uuid_generate_v4()"`
	Name        string    `gorm:"unique"`
	Description string
}
