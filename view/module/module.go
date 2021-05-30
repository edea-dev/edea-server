package module

// SPDX-License-Identifier: EUPL-1.2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/repo"
	"gitlab.com/edea-dev/edea/backend/search"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
	"gorm.io/gorm"
)

// Create a new module
func Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.UserContextKey).(*model.User)

	if err := r.ParseForm(); err != nil {
		view.RenderErrTemplate(r.Context(), w, "module/new.tmpl", err)
		return
	}

	module := new(model.Module)
	if err := util.FormDecoder.Decode(module, r.Form); err != nil {
		view.RenderErrTemplate(r.Context(), w, "module/new.tmpl", err)
		return
	}

	module.ID = uuid.Nil // prevent the client setting an id
	module.UserID = user.ID

	if err := repo.New(module.RepoURL); err != nil && !errors.Is(err, repo.ErrExists) {
		// TODO: display nice error messages
		log.Panic().Err(err).Msg("module: something went wrong fetching the repository")
	}

	result := model.DB.WithContext(r.Context()).Create(module)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not create new module")
	}

	log.Info().Msg("redirecting to new module page")

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		log.Panic().Err(err)
	}

	// redirect to newly created module page
	http.Redirect(w, r, fmt.Sprintf("/module/%s", module.ID), http.StatusSeeOther)
}

// View a module
func View(w http.ResponseWriter, r *http.Request) {
	user, module := getModule(w, r)

	// getModule already writes out the necessary error messages
	if module == nil {
		return
	}

	// get the module author name
	mup := model.Profile{UserID: module.UserID}

	if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
		log.Error().Err(result.Error).Msgf("could not fetch module author profile for user_id %s", module.UserID)
	}

	// render the readme real quick
	g := &repo.Git{URL: module.RepoURL}
	var readme string
	var err error

	if module.Sub != "" {
		readme, err = g.SubModuleReadme(module.Sub)
	} else {
		readme, err = g.Readme()
	}

	if err == nil {
		readme, err = view.RenderReadme(readme)
	} else {
		log.Debug().Err(err).Msg("could not render readme")
	}

	// all packed up,
	m := map[string]interface{}{
		"Module": module,
		"User":   user,
		"Readme": readme,
		"Error":  err,
		"Author": mup.DisplayName,
		"Title":  fmt.Sprintf("EDeA - %s", module.Name),
	}

	// and ready to go
	view.RenderTemplate(r.Context(), "module/view.tmpl", "", m, w)
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
		log.Panic().Err(result.Error).Msg("could not update module")
	}

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		log.Panic().Err(err)
	}

	// redirect to updated module page
	http.Redirect(w, r, fmt.Sprintf("/module/%s", module.ID), http.StatusSeeOther)
}

// Delete a module and redirect to main page
func Delete(w http.ResponseWriter, r *http.Request) {
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

	result := model.DB.WithContext(r.Context()).Delete(&model.Module{ID: uuid.MustParse(moduleID)})
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not delete module")
	}

	// update search index
	if err := search.UpdateEntry(search.Entry{ID: moduleID}); err != nil {
		log.Panic().Err(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// New module form
func New(w http.ResponseWriter, r *http.Request) {
	categories := []model.Category{}

	result := model.DB.Model(&model.Category{}).Find(&categories)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not fetch categories")
	}

	m := map[string]interface{}{
		"Categories": categories,
	}

	view.RenderTemplate(r.Context(), "module/new.tmpl", "EDeA - New Module", m, w)
}

// Pull a module repository
func Pull(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	moduleID := vars["id"]
	ctx := r.Context()

	// check if we even have a module id
	if moduleID == "" {
		msg := map[string]interface{}{
			"Error": "Unfortunately you didn't give us much to work with, try again with a module id.",
		}
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", msg, w)
		return
	}

	user := ctx.Value(util.UserContextKey).(*model.User)

	// try to fetch the module
	module := &model.Module{}

	result := model.DB.Where("id = ? and user_id = ?", moduleID, user.ID).Find(module)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not get the module")
	}

	// nope, no module
	if module.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", nil, w)
		return
	}

	g := &repo.Git{URL: module.RepoURL}
	if err := g.Pull(); err != nil {
		log.Panic().Err(err).Msgf("could not pull latest changes")
	}

	log.Info().Msgf("pulled repo %s for module %s", module.RepoURL, module.ID)

	// redirect to updated module page
	http.Redirect(w, r, fmt.Sprintf("/module/%s", module.ID), http.StatusSeeOther)
}

// ViewHistory provides a commit log of a module
func ViewHistory(w http.ResponseWriter, r *http.Request) {
	user, module := getModule(w, r)

	// getModule already writes out the necessary error messages
	if module == nil {
		return
	}

	// get the module author name
	mup := model.Profile{UserID: module.UserID}

	if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
		log.Error().Err(result.Error).Msgf("could not fetch module author profile for user_id %s", module.UserID)
	}

	// render the readme real quick
	g := &repo.Git{URL: module.RepoURL}
	history, err := g.History(module.Sub)
	if err != nil {
		log.Error().Err(err).Msg("could not get history of repo")
	}

	// all packed up,
	m := map[string]interface{}{
		"Module":  module,
		"User":    user,
		"History": history,
		"Error":   err,
		"Author":  mup.DisplayName,
		"Title":   fmt.Sprintf("EDeA - %s", module.Name),
	}

	// and ready to go
	view.RenderTemplate(r.Context(), "module/view_history.tmpl", "", m, w)
}

// Diff a module's revisions
func Diff(w http.ResponseWriter, r *http.Request) {
	_, module := getModule(w, r)

	// getModule already writes out the necessary error messages
	if module == nil {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Panic().Err(err).Msg("could not parse form data")
	}

	commit1 := r.Form.Get("a")
	commit2 := r.Form.Get("b")

	log.Debug().Msgf("diffing %s and %s", commit1, commit2)

	/*
		# How to do it

		## Schematic
		Call `plotgitsch A B`
		Store the result in the cache

		## Layout
		Fork KiCad-Diff
		Make it just output A and B for each layer
		Get inspired by the existing HTML it comes with, it works already
		Provide a unified view where you can tick/untick layers for visibility

	*/

	// TODO: run diff tools

	// TODO: check out the two revisions from git
}

func plotPCB(mod *model.Module, revision string) ([]byte, error) {
	// processing projects should not take longer than a minute
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	g := &repo.Git{URL: mod.RepoURL}

	// read and parse the module configuration out of the repo
	pcb, err := g.FileByExtAt(mod.Sub, "kicad_pcb", revision)
	if err != nil {
		return nil, util.HintError{
			Hint: fmt.Sprintf("No kicad_pcb file has been found for %s at %s", mod.Sub, revision),
			Err:  err,
		}
	}

	// write the PCB file to disk so we can call kicad via our python script to plot it
	f, err := os.CreateTemp("", revision+".*.kicad_pcb")
	if err != nil {
		log.Panic().Err(err).Msg("could not create temp pcb file")
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(pcb); err != nil {
		log.Panic().Err(err).Msg("could not write temp pcb file contents")
	}

	argv := []string{"plotpcb.py", f.Name()}

	mergeCmd := exec.CommandContext(ctx, "/usr/bin/python3", argv...)

	mergeCmd.Dir = config.Cfg.PlotPCB

	// run the merge
	logOutput, err := mergeCmd.CombinedOutput()

	// return the output of the tool and the error for the user to debug issues
	if err != nil {
		return logOutput, util.HintError{
			Hint: "Something went wrong during the pcb plotting, below is the log which should provide more information.",
			Err:  err,
		}
	}

	return logOutput, nil
}

func getModule(w http.ResponseWriter, r *http.Request) (user *model.User, module *model.Module) {
	vars := mux.Vars(r)
	moduleID := vars["id"]
	ctx := r.Context()

	// check if we even have a module id
	if moduleID == "" {
		msg := map[string]interface{}{
			"Error": "Unfortunately you didn't give us much to work with, try again with a module id.",
		}
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", msg, w)
		return nil, nil
	}

	user, _ = ctx.Value(util.UserContextKey).(*model.User)

	// try to fetch the module
	var result *gorm.DB
	module = new(model.Module)

	if user == nil {
		result = model.DB.Where("id = ? and private = false", moduleID).Preload("Category").Find(module)
	} else {
		result = model.DB.Where("id = ? and (private = false or user_id = ?)", moduleID, user.ID).Preload("Category").Find(module)
	}

	if result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not get the module")
	}

	// nope, no module
	if module.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", nil, w)
		return nil, nil
	}

	return
}
