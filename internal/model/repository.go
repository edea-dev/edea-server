package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"time"

	"github.com/google/uuid"
)

// Repository model for the repo cache
type Repository struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	URL       string    // fetch/clone URL
	Type      string    // VCS type (e.g. git)
	Location  string    // filesystem path
	UpdatedAt time.Time // time of last fetch
	CreatedAt time.Time // entry added
}
