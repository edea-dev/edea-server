package model

import (
	"github.com/rs/zerolog"
)

// BenchModule contains the configuration for a Module as part of a Bench
type BenchModule struct {
	ID          string `pg:"type:uuid,default:gen_random_uuid(),pk"`
	Name        string
	Description string
	Conf        map[string]interface{}
	ModuleID    string  `pg:"type:uuid,notnull"`
	Module      *Module `pg:"rel:has-one"`
	BenchID     string  `pg:"type:uuid,notnull"`
	Bench       *Bench  `pg:"rel:has-one"`
}

// MarshalZerologObject provides the object representation for logging
func (bp *BenchModule) MarshalZerologObject(e *zerolog.Event) {
	e.Str("bench_module_uuid", bp.ID)
}
