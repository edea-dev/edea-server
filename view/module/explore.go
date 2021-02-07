package module

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/view"
)

// TOOD: improve DDL so we can do this with the ORM natively
type ExploreModule struct {
	ID          string
	UserID      string
	DisplayName string
	Name        string
	Description string
	RepoURL     string
}

const exploreQuery = `
	SELECT m.id, m.user_id, p.display_name, m.name, m.description, m.repo_url
	FROM modules m
	JOIN profiles p
		ON p.user_id = m.user_id
	WHERE m.private = false
		AND m.deleted_at IS NULL
	ORDER BY m.updated_at;`

// Explore modules page
func Explore(w http.ResponseWriter, r *http.Request) {
	var p []ExploreModule

	user := view.CurrentUser(r)

	result := model.DB.Raw(exploreQuery).Scan(&p)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not run explore query")
	}

	m := map[string]interface{}{
		"Modules": p,
		"User":    user,
	}

	view.RenderTemplate("explore/view.tmpl", m, w)
}

// ExploreUser lists all the public modules by a specific user
func ExploreUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		log.Panic().Msg("explore_user: no user id specified")
	}

	currentUser := view.CurrentUser(r)
	var id uuid.UUID

	// /explore/user/me should return your own modules
	if userID == "me" {
		// redirect to the login page if someone isn't logged in
		if currentUser == nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		id = currentUser.ID
	} else {
		id = uuid.MustParse(userID)
	}

	p := &model.Profile{UserID: id}

	result := model.DB.Find(&p)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("no such user")
	}

	var modules []model.Module
	m := make(map[string]interface{})

	if currentUser != nil && id == currentUser.ID {
		// show a users own modules
		result = model.DB.Where(&model.Module{UserID: id}).Find(&modules)
		m["Self"] = true
	} else {
		// show another users modules
		result = model.DB.Where("user_id = ? AND private = false AND deleted_at IS NULL", id).Find(&modules)
	}
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not run user explore query")
	}

	m["Modules"] = modules
	m["Profile"] = p

	view.RenderTemplate("explore/user.tmpl", m, w)
}
