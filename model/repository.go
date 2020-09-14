package model

import "time"

// Repository model for the repo cache
type Repository struct {
	ID       string    `pg:"type:uuid,default:gen_random_uuid(),pk"`
	URL      string    // fetch/clone URL
	Type     string    // VCS type (e.g. git)
	Location string    // filesystem path
	Updated  time.Time // time of last fetch
	Created  time.Time // entry added
}
