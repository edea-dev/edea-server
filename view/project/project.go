package project

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

// Explore projects page
func Explore(w http.ResponseWriter, r *http.Request) {
	var p []model.Project

	err := model.DB.Model(&p).
		OrderExpr("project.uuid ASC").
		Limit(10).
		Select()
	if err != nil {
		log.Panic().Err(err).Msg("could not fetch projects")
	}

	m := map[string]interface{}{
		"Projects": p,
	}

	log.Printf("%+v", m)

	view.RenderMarkdown("explore.md", m, w)
}

// New Project view
func New(w http.ResponseWriter, r *http.Request) {
	user := view.CurrentUser(r)

	if user == nil {
		log.Panic().Err(util.ErrImSorryDave).Msg("no can do")
	}

	data := map[string]interface{}{
		"User": user,
	}

	view.RenderMarkdown("project/new.md", data, w)
}

// Create a new project
func Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "project/new.md", err)
		return
	}

	project := new(model.Project)
	if err := util.FormDecoder.Decode(project, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "project/new.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	if project.UserID != user.ID {
		if user.IsAdmin {
			// TODO: admin changing stuff
		} else {
			view.RenderErr(r.Context(), w, "project/new.md", util.ErrImSorryDave)
			return
		}
	}

	project.ID = "" // prevent the client setting an id

	_, err := model.DB.Model(project).Update()
	if err != nil {
		log.Panic().Err(err).Msg("could not update profile")
	}

	// redirect to newly created project page
	http.Redirect(w, r, fmt.Sprintf("/project/%s", project.ID), http.StatusSeeOther)
}

// View a project
func View(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	// check if we even have a project id
	if projectID == "" {
		msg := map[string]interface{}{
			"Error": "Unfortunately you didn't give us much to work with, try again with a project id.",
		}
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("project/404.md", msg, w)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	// try to fetch the project
	project := &model.Project{}

	err := model.DB.Model(project).Where("id = ? and (private = false or user_id = ?)", projectID, user.ID).Select()
	if err != nil {
		log.Panic().Err(err).Msgf("could not get the project: %v", err)
	}

	// nope, no project
	if project.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("project/404.md", nil, w)
		return
	}

	// render the readme real quick
	g := &repo.Git{URL: project.RepoURL}
	readme, err := g.Readme()

	if err == nil {
		readme, err = view.RenderReadme(readme)
	}

	// all packed up,
	m := map[string]interface{}{
		"Project": project,
		"User":    user,
		"Readme":  readme,
		"Error":   err,
	}

	// and ready to go
	view.RenderMarkdown("project/view.md", m, w)
}

// Update a project and reload the page
func Update(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "project/view.md", err)
		return
	}

	project := new(model.Project)
	if err := util.FormDecoder.Decode(project, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "project/view.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	if project.UserID != user.ID {
		if user.IsAdmin {
			// TODO: admin changing stuff
		} else {
			view.RenderErr(r.Context(), w, "user/profile.md", util.ErrImSorryDave)
			return
		}
	}

	_, err := model.DB.Model(project).Update()
	if err != nil {
		log.Panic().Err(err).Msg("could not update profile")
	}

	// redirect to updated project page
	http.Redirect(w, r, fmt.Sprintf("/project/%s", project.ID), http.StatusSeeOther)
}
