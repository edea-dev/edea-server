package model

// SPDX-License-Identifier: EUPL-1.2

// Category model
type Category struct {
	ID          string `gorm:"type:uuid;primarykey;default:uuid_generate_v4()"`
	Name        string `gorm:"unique"`
	Description string
}
