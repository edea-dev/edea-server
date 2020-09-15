package module

import (
	"fmt"
	"net/http"

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

	err := model.DB.Model(&p).
		OrderExpr("module.uuid ASC").
		Limit(10).
		Select()
	if err != nil {
		log.Panic().Err(err).Msg("could not fetch modules")
	}

	m := map[string]interface{}{
		"Modules": p,
	}

	log.Printf("%+v", m)

	view.RenderMarkdown("explore.md", m, w)
}

// New Module view
func New(w http.ResponseWriter, r *http.Request) {
	user := view.CurrentUser(r)

	if user == nil {
		log.Panic().Err(util.ErrImSorryDave).Msg("no can do")
	}

	data := map[string]interface{}{
		"User": user,
	}

	view.RenderMarkdown("module/new.md", data, w)
}

// Create a new module
func Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "module/new.md", err)
		return
	}

	module := new(model.Module)
	if err := util.FormDecoder.Decode(module, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "module/new.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	if module.UserID != user.ID {
		if user.IsAdmin {
			// TODO: admin changing stuff
		} else {
			view.RenderErr(r.Context(), w, "module/new.md", util.ErrImSorryDave)
			return
		}
	}

	module.ID = "" // prevent the client setting an id

	_, err := model.DB.Model(module).Insert()
	if err != nil {
		log.Panic().Err(err).Msg("could not update profile")
	}

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

	user := r.Context().Value(util.UserContextKey).(model.User)

	// try to fetch the module
	module := &model.Module{}

	err := model.DB.Model(module).Where("id = ? and (private = false or user_id = ?)", moduleID, user.ID).Select()
	if err != nil {
		log.Panic().Err(err).Msgf("could not get the module: %v", err)
	}

	// nope, no module
	if module.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", nil, w)
		return
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
	}

	// and ready to go
	view.RenderMarkdown("module/view.md", m, w)
}

// Update a module and reload the page
func Update(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "module/view.md", err)
		return
	}

	module := new(model.Module)
	if err := util.FormDecoder.Decode(module, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "module/view.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	if module.UserID != user.ID {
		if user.IsAdmin {
			// TODO: admin changing stuff
		} else {
			view.RenderErr(r.Context(), w, "user/profile.md", util.ErrImSorryDave)
			return
		}
	}

	_, err := model.DB.Model(module).Update()
	if err != nil {
		log.Panic().Err(err).Msg("could not update profile")
	}

	// redirect to updated module page
	http.Redirect(w, r, fmt.Sprintf("/module/%s", module.ID), http.StatusSeeOther)
}
