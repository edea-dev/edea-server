package module

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/model"
	"gitlab.com/edea-dev/edead/util"
	"gitlab.com/edea-dev/edead/view"
)

// TOOD: improve DDL so we can do this with the ORM natively

type ExploreModule struct {
	ID          string
	UserID      string
	DisplayName string
	Name        string
	Description string
	RepoURL     string
	Category    string
}

const exploreQuery = `
	SELECT m.id, m.user_id, p.display_name, m.name, m.description, m.repo_url, c.name as category
	FROM modules m
	JOIN profiles p
		ON p.user_id = m.user_id
	JOIN categories c
		ON c.id = m.category_id
	WHERE m.private = false
		AND m.deleted_at IS NULL
	ORDER BY m.updated_at;`

// Explore modules page
func Explore(w http.ResponseWriter, r *http.Request) {
	var p []ExploreModule

	result := model.DB.Raw(exploreQuery).Scan(&p)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not run explore query")
	}

	m := map[string]interface{}{
		"Modules": p,
	}

	view.RenderTemplate(r.Context(), "explore/view.tmpl", "EDeA - Explore Modules", m, w)
}

// ExploreUser lists all the public modules by a specific user
func ExploreUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	ctx := r.Context()
	currentUser, _ := ctx.Value(util.UserContextKey).(*model.User)

	if userID == "" {
		log.Panic().Msg("explore_user: no user id specified")
	}

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

	view.RenderTemplate(ctx, "explore/user.tmpl", "EDeA - User Modules", m, w)
}
