package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
	"gorm.io/datatypes"
)

// BenchModule contains the configuration for a Module as part of a Bench
type BenchModule struct {
	ID          uuid.UUID `gorm:"type:uuid;primarykey;default:uuid_generate_v4()"`
	Name        string
	Description string
	Conf        datatypes.JSON
	ModuleID    uuid.UUID `gorm:"type:uuid"`
	Module      Module
	BenchID     uuid.UUID `gorm:"type:uuid"`
	Bench       Bench

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
}

// MarshalLogObject provides the object representation for logging
func (bm *BenchModule) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("bench_module_uuid", bm.ID.String())

	return nil
}
