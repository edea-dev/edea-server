package model

import (
	"github.com/rs/zerolog"
)

// BenchProject contains the configuration for a Project as part of a Bench
type BenchProject struct {
	ID          string `pg:"type:uuid,default:gen_random_uuid(),pk"`
	Name        string
	Description string
	Conf        map[string]interface{}
	ProjectID   string   `pg:"type:uuid,notnull"`
	Project     *Project `pg:"rel:has-one"`
	BenchID     string   `pg:"type:uuid,notnull"`
	Bench       *Bench   `pg:"rel:has-one"`
}

// MarshalZerologObject provides the object representation for logging
func (bp *BenchProject) MarshalZerologObject(e *zerolog.Event) {
	e.Str("bench_project_uuid", bp.ID)
}
