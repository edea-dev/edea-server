package model

import (
	"github.com/rs/zerolog"
)

// Project model
type Project struct {
	ID          string `pg:"type:uuid,default:gen_random_uuid(),pk"`
	UserID      string `pg:"type:uuid,fk"`
	User        *User  `pg:"rel:has-one"`
	RepoURL     string `schema:"repourl,required"`
	Name        string `schema:"name,required"`
	Description string `schema:"description"`
}

// MarshalZerologObject provides the object representation for logging
func (p *Project) MarshalZerologObject(e *zerolog.Event) {
	e.Str("project_uuid", p.ID)
}
