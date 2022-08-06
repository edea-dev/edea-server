package model

import "gorm.io/gorm"

// SPDX-License-Identifier: EUPL-1.2

// Filter model
type Filter struct {
	gorm.Model
	Key         string `gorm:"unique"`
	Name        string `gorm:"unique"`
	Description string
}
