package model

import (
	"github.com/rs/zerolog"
)

// Profile data for Users
type Profile struct {
	ID          string `pg:"type:uuid,default:gen_random_uuid(),pk"`
	UserID      string `pg:"type:uuid,fk"`
	DisplayName string `schema:"display_name,required"`
	Location    string
	Biography   string
	Avatar      string
}

// MarshalZerologObject marshaller to log profile objects
func (p *Profile) MarshalZerologObject(e *zerolog.Event) {
	e.Str("profile_uuid", p.ID)
}
