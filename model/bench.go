package model

import (
	"time"

	"github.com/rs/zerolog"
)

// Bench contains a number of projects with their configuration
type Bench struct {
	ID          string `pg:"type:uuid,default:gen_random_uuid(),pk"`
	UserID      string `pg:"type:uuid,fk"`
	User        *User  `pg:"rel:has-one"`
	Active      bool   // i.e. only show current active bench
	Public      bool
	Created     time.Time       `pg:",default:now()"` // creation date, automatically set to now
	Projects    []*BenchProject `pg:"many2many:bench_projects,join_fk:id"`
	Name        string
	Description string
}

// MarshalZerologObject provides the object representation for logging
func (b *Bench) MarshalZerologObject(e *zerolog.Event) {
	e.Str("bench_uuid", b.ID)
}
