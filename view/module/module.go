package module

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/repo"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
)

// Explore modules page
func Explore(w http.ResponseWriter, r *http.Request) {
	var p []model.Module

	result := model.DB.Order("id ASC").Limit(10).Find(&p)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not fetch modules")
	}

	m := map[string]interface{}{
		"Modules": p,
	}

	view.RenderTemplate("explore.tmpl", m, w)
}

// New Module view
func New(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.UserContextKey)

	if user == nil {
		log.Panic().Err(util.ErrImSorryDave).Msg("user not logged in")
	}

	data := map[string]interface{}{
		"User": user,
	}

	view.RenderMarkdown("module/new.md", data, w)
}

// Create a new module
func Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.UserContextKey).(*model.User)

	if err := r.ParseForm(); err != nil {
		view.RenderErrMarkdown(r.Context(), w, "module/new.md", err)
		return
	}

	module := new(model.Module)
	if err := util.FormDecoder.Decode(module, r.Form); err != nil {
		view.RenderErrMarkdown(r.Context(), w, "module/new.md", err)
		return
	}

	module.ID = uuid.Nil // prevent the client setting an id
	module.UserID = user.ID

	result := model.DB.WithContext(r.Context()).Create(module)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not create new module")
	}

	log.Info().Msg("redirecting to new module page")

	// redirect to newly created module page
	http.Redirect(w, r, fmt.Sprintf("/module/%s", module.ID), http.StatusSeeOther)
}

// View a module
func View(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	moduleID := vars["id"]

	// check if we even have a module id
	if moduleID == "" {
		msg := map[string]interface{}{
			"Error": "Unfortunately you didn't give us much to work with, try again with a module id.",
		}
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", msg, w)
		return
	}

	user := r.Context().Value(util.UserContextKey).(*model.User)

	// try to fetch the module
	module := &model.Module{}

	result := model.DB.Where("id = ? and (private = false or user_id = ?)", moduleID, user.ID).Find(module)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not get the module")
	}

	// nope, no module
	if module.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", nil, w)
		return
	}

	// get the module author name
	mup := model.Profile{UserID: module.UserID}

	if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
		log.Error().Err(result.Error).Msgf("could not fetch module author profile for user_id %s", module.UserID)
	}

	// render the readme real quick
	g := &repo.Git{URL: module.RepoURL}
	readme, err := g.Readme()

	if err == nil {
		readme, err = view.RenderReadme(readme)
	}

	// all packed up,
	m := map[string]interface{}{
		"Module": module,
		"User":   user,
		"Readme": readme,
		"Error":  err,
		"Author": mup.DisplayName,
	}

	// and ready to go
	view.RenderMarkdown("module/view.md", m, w)
}

// Update a module and reload the page
func Update(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErrMarkdown(r.Context(), w, "module/view.md", err)
		return
	}

	module := new(model.Module)
	if err := util.FormDecoder.Decode(module, r.Form); err != nil {
		view.RenderErrMarkdown(r.Context(), w, "module/view.md", err)
		return
	}

	result := model.DB.WithContext(r.Context()).Save(module)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not update profile")
	}

	// redirect to updated module page
	http.Redirect(w, r, fmt.Sprintf("/module/%s", module.ID), http.StatusSeeOther)
}
