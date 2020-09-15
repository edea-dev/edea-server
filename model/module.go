package model

import (
	"github.com/rs/zerolog"
)

// Module model
type Module struct {
	ID          string `pg:"type:uuid,default:gen_random_uuid(),pk"`
	UserID      string `pg:"type:uuid,fk"`
	User        *User  `pg:"rel:has-one"`
	RepoURL     string `schema:"repourl,required"`
	Name        string `schema:"name,required"`
	Description string `schema:"description"`
}

// MarshalZerologObject provides the object representation for logging
func (p *Module) MarshalZerologObject(e *zerolog.Event) {
	e.Str("module_uuid", p.ID)
}
