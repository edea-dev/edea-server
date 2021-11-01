package module

// SPDX-License-Identifier: EUPL-1.2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/edea-dev/edead/internal/config"
	"gitlab.com/edea-dev/edead/internal/merge"
	"gitlab.com/edea-dev/edead/internal/model"
	"gitlab.com/edea-dev/edead/internal/repo"
	"gitlab.com/edea-dev/edead/internal/search"
	"gitlab.com/edea-dev/edead/internal/util"
	"gitlab.com/edea-dev/edead/internal/view"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Board as we get it from the diff tool
type Board struct {
	X1     float32           `json:"x1"`
	X2     float32           `json:"x2"`
	Width  float32           `json:"width"`
	Height float32           `json:"height"`
	Layers map[string]string `json:"layers"`
}

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
		zap.L().Panic("module: something went wrong fetching the repository", zap.Error(err))
	}

	meta, err := merge.Metadata(module)
	if err != nil {
		zap.L().Panic("metadata extraction unsuccessful", zap.Error(err))
	}

	b, _ := json.Marshal(meta)
	module.Metadata = b

	result := model.DB.WithContext(r.Context()).Create(module)
	if result.Error != nil {
		zap.L().Panic("could not create new module", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
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
		zap.L().Error("could not fetch module author profile", zap.Error(result.Error), zap.String("user_id", module.UserID.String()))
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
	}
	if err != nil {
		zap.L().Debug("could not render readme", zap.Error(err))
	}

	hasDocs, err := g.HasDocs(module.Sub)
	if err != nil {
		hasDocs = false
	}

	// all packed up,
	m := map[string]interface{}{
		"Module":  module,
		"User":    user,
		"Readme":  readme,
		"Error":   err,
		"Author":  mup.DisplayName,
		"HasDocs": hasDocs,
		"Title":   fmt.Sprintf("EDeA - %s", module.Name),
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

	meta, err := merge.Metadata(module)
	if err != nil {
		view.RenderErrMarkdown(r.Context(), w, "module/view.md", err)
		return
	}

	if err := module.Metadata.Scan(meta); err != nil {
		view.RenderErrMarkdown(r.Context(), w, "module/view.md", err)
		return
	}

	result := model.DB.WithContext(r.Context()).Save(module)
	if result.Error != nil {
		zap.L().Panic("could not update module", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
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
		zap.L().Panic("could not delete module", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.Entry{ID: moduleID}); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// New module form
func New(w http.ResponseWriter, r *http.Request) {
	categories := []model.Category{}

	result := model.DB.Model(&model.Category{}).Find(&categories)
	if result.Error != nil {
		zap.L().Panic("could not fetch categories", zap.Error(result.Error))
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
		view.RenderErrTemplate(ctx, w, "module/404.tmpl", errors.New("Unfortunately you didn't give us much to work with, try again with a module id"))
		return
	}

	user := ctx.Value(util.UserContextKey).(*model.User)

	// try to fetch the module
	module := &model.Module{}

	result := model.DB.Where("id = ? and user_id = ?", moduleID, user.ID).Find(module)
	if result.Error != nil {
		zap.L().Panic("could not get the module", zap.Error(result.Error))
	}

	// nope, no module
	if module.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		view.RenderErrTemplate(ctx, w, "module/404.md", errors.New("No such Module"))
		return
	}

	g := &repo.Git{URL: module.RepoURL}
	if err := g.Pull(); err != nil {
		zap.L().Panic("could not pull latest changes", zap.Error(err))
	}

	meta, err := merge.Metadata(module)
	if err != nil {
		zap.L().Error("metadata extraction unsuccessful", zap.Error(err))
		view.RenderErrTemplate(r.Context(), w, "module/view.tmpl", err)
		return
	}

	b, _ := json.Marshal(meta)

	module.Metadata = b

	result = model.DB.WithContext(r.Context()).Save(module)
	if result.Error != nil {
		zap.L().Panic("could not update submodule", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	zap.L().Info("pulled repo for module", zap.String("repo_url", module.RepoURL), zap.String("module_id", module.ID.String()))

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
		zap.L().Error("could not fetch module author profile", zap.Error(result.Error), zap.String("user_id", module.UserID.String()))
	}

	// render the readme real quick
	g := &repo.Git{URL: module.RepoURL}
	history, err := g.History(module.Sub)
	if err != nil {
		zap.L().Error("could not get history of repo", zap.Error(err))
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
		zap.L().Panic("could not parse form data", zap.Error(err))
	}

	commit1 := r.Form.Get("a")
	commit2 := r.Form.Get("b")

	zap.S().Debugf("diffing %s and %s", commit1, commit2)

	pcba, err := plotPCB(module, commit1)
	if err != nil {
		zap.L().Panic("could not plot pcb at A", zap.Error(err), zap.String("commit", commit1))
	}
	pcbb, err := plotPCB(module, commit1)
	if err != nil {
		zap.L().Panic("could not plot pcb at B", zap.Error(err), zap.String("commit", commit2))
	}

	g := &repo.Git{URL: module.RepoURL}

	scha, err := g.SchematicHelper(module.Sub, commit1)
	if err != nil {
		zap.L().Panic("failed to plot sch a", zap.Error(err))
	}

	schb, err := g.SchematicHelper(module.Sub, commit2)
	if err != nil {
		zap.L().Panic("failed to plot sch b", zap.Error(err))
	}

	m := map[string]interface{}{
		"Module": module,
		"PCBA":   pcba,
		"PCBB":   pcbb,
		"SCHA":   scha,
		"SCHB":   schb,
		"Title":  fmt.Sprintf("EDeA - %s", module.Name),
	}

	// and ready to go
	view.RenderTemplate(r.Context(), "module/view_diff.tmpl", "", m, w)
}

func plotPCB(mod *model.Module, revision string) (*Board, error) {
	// processing projects should not take longer than a minute
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	g := &repo.Git{URL: mod.RepoURL}

	// read and parse the module configuration out of the repo
	pcb, err := g.FileByExtAt(mod.Sub, ".kicad_pcb", revision)
	if err != nil {
		return nil, util.HintError{
			Hint: fmt.Sprintf("No kicad_pcb file has been found for %s at %s", mod.Sub, revision),
			Err:  err,
		}
	}

	// write the PCB file to disk so we can call kicad via our python script to plot it
	f, err := os.CreateTemp("", revision+".*.kicad_pcb")
	if err != nil {
		zap.L().Panic("could not create temp pcb file", zap.Error(err))
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(pcb); err != nil {
		zap.L().Panic("could not write temp pcb file contents", zap.Error(err))
	}

	argv := []string{config.Cfg.Tools.PlotPCB, f.Name()}

	plotCmd := exec.CommandContext(ctx, "python3", argv...)

	// run the plotting operation
	logOutput, err := plotCmd.CombinedOutput()

	// return the output of the tool and the error for the user to debug issues
	if err != nil {
		zap.L().Info("plot pcb output", zap.ByteString("output", logOutput))
		return nil, util.HintError{
			Hint: fmt.Sprintf("Something went wrong during the pcb plotting, below is the log which should provide more information:\n%s", logOutput),
			Err:  err,
		}
	}

	b := new(Board)
	json.Unmarshal(logOutput, b)

	return b, nil
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
		zap.L().Panic("could not get the module", zap.Error(result.Error))
	}

	// nope, no module
	if module.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("module/404.md", nil, w)
		return nil, nil
	}

	return
}

// BuildBook runs mdbook on the /doc (or otherwise configured) folder of the module to generate documentation
func BuildBook(w http.ResponseWriter, r *http.Request) {
	_, module := getModule(w, r)

	// getModule already writes out the necessary error messages
	if module == nil {
		return
	}

	// render the readme real quick
	g := &repo.Git{URL: module.RepoURL}
	var docPath string
	var err error

	if module.Sub != "" {
		docPath, err = g.SubModuleDocs(module.Sub)
	}

	repoPath, err := g.Dir()

	s := filepath.Join(repoPath, docPath)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dest := filepath.Join(config.Cfg.Cache.Book.Base, module.ID.String())

	zap.L().Debug("book destination", zap.String("path", dest))

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err := os.Mkdir(dest, 0755)
		if err != nil {
			zap.L().Panic("could not create book dir", zap.Error(err))
		}
	}

	argv := []string{"build", s, "-d", dest}

	bookCmd := exec.CommandContext(ctx, "mdbook", argv...)

	// run the merge
	logOutput, err := bookCmd.CombinedOutput()

	// show the user the tool output in case of an error while building the book
	if err != nil {
		m := map[string]interface{}{
			"Error":  err,
			"Output": strings.ReplaceAll(string(logOutput), "\n", "<br>"),
		}
		if err, ok := err.(util.HintError); ok {
			m["Error"] = err
			m["Hint"] = "Something went wrong during building the book, please see the log"
		}
		view.RenderTemplate(ctx, "bench/merge_error.tmpl", "mdbook Error", m, w)
		return
	}

	zap.L().Debug("build book log output", zap.ByteString("output", logOutput))

	http.Redirect(w, r, fmt.Sprintf("/module/doc/%s", module.ID), http.StatusTemporaryRedirect)
}

func PullAllRepos(w http.ResponseWriter, r *http.Request) {
	var modules []model.Module

	result := model.DB.Find(&modules)
	if result.Error != nil {
		zap.L().Panic("could not fetch all modules", zap.Error(result.Error))
	}

	for _, mod := range modules {
		g := &repo.Git{URL: mod.RepoURL}
		repo.Add(mod.RepoURL)
		if err := g.Pull(); err != nil {
			zap.L().Error("could not pull latest changes", zap.Error(err))
			continue
		}

		meta, err := merge.Metadata(&mod)
		if err != nil {
			zap.L().Error("metadata extraction unsuccessful", zap.Error(err))
			continue
		}

		b, _ := json.Marshal(meta)

		mod.Metadata = b

		result = model.DB.WithContext(r.Context()).Save(mod)
		if result.Error != nil {
			zap.L().Error("could not update module", zap.Error(err))
			continue
		}

		// update search index
		if err := search.UpdateEntry(search.ModuleToEntry(mod)); err != nil {
			zap.L().Error("could not update search index", zap.Error(err))
			continue
		}
	}
}
