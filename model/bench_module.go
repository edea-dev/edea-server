package model

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BenchModule contains the configuration for a Module as part of a Bench
type BenchModule struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name        string
	Description string
	Conf        datatypes.JSON
	ModuleID    uuid.UUID `gorm:"type:uuid"`
	Module      Module
	BenchID     uuid.UUID `gorm:"type:uuid"`
	Bench       Bench
}

// MarshalZerologObject provides the object representation for logging
func (bm *BenchModule) MarshalZerologObject(e *zerolog.Event) {
	e.Str("bench_module_uuid", bm.ID.String())
}
