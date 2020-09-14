package model

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO: handle the case where we might want to link multiple users
// to a single identity. e.g. to migrate accounts or to use OIDC
// and WebAuth simultanously.

// User mapping from IDs to the authentication provider data
type User struct {
	ID        string    `pg:"type:uuid,default:gen_random_uuid(),pk"`
	AuthUUID  string    `pg:",unique,notnull"` // unique id from authentication provider
	Handle    string    `pg:",unique"`         // user handle as it will be used in the url
	Created   time.Time `pg:",default:now()"`  // creation date, automatically set to now
	IsAdmin   bool      `pg:"type:boolean,default:false"`
	ProfileID string    `pg:"type:uuid,notnull"`
	Profile   *Profile  `pg:"rel:has-one"`
}

// MarshalZerologObject provides the object representation for logging
func (u *User) MarshalZerologObject(e *zerolog.Event) {
	e.Str("uuid", u.ID).
		Str("auth_uuid", u.AuthUUID).
		Str("handle", u.Handle).
		Time("created", u.Created)
}

// UserExists returns true if a user with the given auth uuid exists
func UserExists(authUUID string) bool {
	u := User{AuthUUID: authUUID}

	if err := DB.Model(&u).Select(); err != nil && err != pg.ErrNoRows {
		log.Panic().Err(err).Msgf("could not get user")
	}

	if u.ID == "" {
		return false
	}
	return true
}
