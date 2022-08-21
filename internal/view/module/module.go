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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/edea-dev/edea-server/internal/config"
	"gitlab.com/edea-dev/edea-server/internal/merge"
	"gitlab.com/edea-dev/edea-server/internal/model"
	"gitlab.com/edea-dev/edea-server/internal/repo"
	"gitlab.com/edea-dev/edea-server/internal/search"
	"gitlab.com/edea-dev/edea-server/internal/util"
	"gitlab.com/edea-dev/edea-server/internal/view"
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
func Create(c *gin.Context) {
	user := c.Keys["user"].(*model.User)

	module := new(model.Module)

	if err := c.Bind(module); err != nil {
		view.RenderErrTemplate(c, "module/new.tmpl", err)
		return
	}

	// check if it already exists and redirect to it if it does
	tm := model.Module{RepoURL: module.RepoURL, Sub: module.Sub}
	result := model.DB.Where(&tm).First(&tm)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			zap.S().Panic(result.Error)
		}
	} else {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/module/%s", tm.ID))
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
		zap.S().Panic(err)
	}

	module.Metadata = meta

	result = model.DB.WithContext(c).Create(module)
	if result.Error != nil {
		zap.L().Panic("could not create new module", zap.Error(result.Error))
	}

	// get the full object from the database to update it in meilisearch
	result = model.DB.WithContext(c).Preload("User").Preload("Category").First(module)
	if result.Error != nil {
		zap.L().Panic("could not create load new module", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	// redirect to newly created module page
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/module/%s", module.ID))
}

// View a module
func View(c *gin.Context) {
	user, module := getModule(c)
	ref := c.Query("ref")
	if ref == "" {
		ref = "HEAD"
	}

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
		readme, err = g.SubModuleReadme(module.Sub, ref)
	} else {
		readme, err = g.Readme(ref)
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
	view.RenderTemplate(c, "module/view.tmpl", "", m)
}

// View a module
func UpdateView(c *gin.Context) {
	categories := []model.Category{}
	user, module := getModule(c)

	// getModule already writes out the necessary error messages
	if module == nil {
		return
	}

	// get the module author name
	mup := model.Profile{UserID: module.UserID}

	if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
		zap.L().Error("could not fetch module author profile", zap.Error(result.Error), zap.String("user_id", module.UserID.String()))
	}

	result := model.DB.Model(&model.Category{}).Find(&categories)
	if result.Error != nil {
		zap.L().Panic("could not fetch categories", zap.Error(result.Error))
	}

	// all packed up,
	m := map[string]interface{}{
		"Module":     module,
		"User":       user,
		"Error":      nil,
		"Author":     mup.DisplayName,
		"Title":      fmt.Sprintf("EDeA - %s", module.Name),
		"Categories": categories,
	}

	// and ready to go
	view.RenderTemplate(c, "module/update.tmpl", "", m)
}

// Update a module and reload the page
func Update(c *gin.Context) {
	var tm model.Module
	var module = new(model.Module)
	moduleID := uuid.MustParse(c.Param("id"))

	if err := c.Bind(module); err != nil {
		view.RenderErrTemplate(c, "module/update.tmpl", err)
		return
	}

	result := model.DB.First(&tm, moduleID)
	if result.Error != nil {
		zap.S().Panic(result.Error)
	}

	tm.Name = module.Name
	tm.Description = module.Description
	tm.Private = module.Private
	tm.CategoryID = module.CategoryID

	result = model.DB.WithContext(c).Save(&tm)
	if result.Error != nil {
		zap.L().Panic("could not update module", zap.Error(result.Error))
	}

	// get the full object from the database to update it in meilisearch
	result = model.DB.WithContext(c).Preload("User").Preload("Category").First(module)
	if result.Error != nil {
		zap.L().Panic("could not create load new module", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	// redirect to updated module page
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/module/%s", module.ID))
}

// Delete a module and redirect to main page
func Delete(c *gin.Context) {
	moduleID := c.Param("id")

	// check if we even have a module id
	if moduleID == "" {
		msg := map[string]interface{}{
			"Error": "Unfortunately you didn't give us much to work with, try again with a module id.",
		}
		c.Status(http.StatusNotFound)
		view.RenderTemplate(c, "module/404.tmpl", "Module not found", msg)
		return
	}

	result := model.DB.WithContext(c).Delete(&model.Module{ID: uuid.MustParse(moduleID)})
	if result.Error != nil {
		zap.L().Panic("could not delete module", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.Entry{ID: moduleID}); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	c.Redirect(http.StatusSeeOther, "/")
}

// New module form
func New(c *gin.Context) {
	categories := []model.Category{}

	result := model.DB.Model(&model.Category{}).Find(&categories)
	if result.Error != nil {
		zap.L().Panic("could not fetch categories", zap.Error(result.Error))
	}

	m := map[string]interface{}{
		"Categories": categories,
	}

	view.RenderTemplate(c, "module/new.tmpl", "EDeA - New Module", m)
}

// Pull a module repository
func Pull(c *gin.Context) {
	moduleID := c.Param("id")

	// check if we even have a module id
	if moduleID == "" {
		view.RenderErrTemplate(c, "module/404.tmpl", errors.New("Unfortunately you didn't give us much to work with, try again with a module id"))
		return
	}

	user := c.Keys["user"].(*model.User)

	// try to fetch the module
	module := &model.Module{}

	result := model.DB.Where("id = ? and user_id = ?", moduleID, user.ID).Find(module)
	if result.Error != nil {
		zap.L().Panic("could not get the module", zap.Error(result.Error))
	}

	// nope, no module
	if module.ID == uuid.Nil {
		c.Status(http.StatusNotFound)
		view.RenderErrTemplate(c, "module/404.md", errors.New("No such Module"))
		return
	}

	g := &repo.Git{URL: module.RepoURL}
	if err := g.Pull(); err != nil {
		zap.L().Panic("could not pull latest changes", zap.Error(err))
	}

	meta, err := merge.Metadata(module)
	if err != nil {
		zap.L().Error("metadata extraction unsuccessful", zap.Error(err))
		view.RenderErrTemplate(c, "module/view.tmpl", err)
		return
	}

	module.Metadata = meta

	result = model.DB.WithContext(c).Save(module)
	if result.Error != nil {
		zap.L().Panic("could not update submodule", zap.Error(result.Error))
	}

	// get the full object from the database to update it in meilisearch
	result = model.DB.WithContext(c).Preload("User").Preload("Category").First(module)
	if result.Error != nil {
		zap.L().Panic("could not create load new module", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.ModuleToEntry(*module)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	zap.L().Info("pulled repo for module", zap.String("repo_url", module.RepoURL), zap.String("module_id", module.ID.String()))

	// redirect to updated module page
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/module/%s", module.ID))
}

// ViewHistory provides a commit log of a module
func ViewHistory(c *gin.Context) {
	user, module := getModule(c)

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
	view.RenderTemplate(c, "module/view_history.tmpl", "", m)
}

// Diff a module's revisions
func Diff(c *gin.Context) {
	_, module := getModule(c)

	// getModule already writes out the necessary error messages
	if module == nil {
		return
	}

	commit1 := c.Query("a")
	commit2 := c.Query("b")

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
	view.RenderTemplate(c, "module/view_diff.tmpl", "", m)
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

	argv := []string{"plotpcb", f.Name()}

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

func getModule(c *gin.Context) (user *model.User, module *model.Module) {
	moduleID := c.Param("id")

	// check if we even have a module id
	if moduleID == "" {
		msg := map[string]interface{}{
			"Error": "Unfortunately you didn't give us much to work with, try again with a module id.",
		}
		c.Status(http.StatusNotFound)
		view.RenderTemplate(c, "module/404.tmpl", "Module not found", msg)
		return nil, nil
	}

	user, _ = c.Keys["user"].(*model.User)

	// try to fetch the module
	var result *gorm.DB
	module = new(model.Module)

	if user == nil {
		result = model.DB.Where("id = ? and private = false", moduleID).Preload("Category").Find(module)
	} else {
		result = model.DB.Where("id = ? and (private = false or user_id = ?)", moduleID, user.ID).Preload("Category").Find(module)
	}

	if result.Error != nil {
		zap.L().Error("could not get the module", zap.Error(result.Error))
		c.Status(http.StatusNotFound)
		view.RenderErrTemplate(c, "module/404.tmpl", result.Error)
		return nil, nil
	}

	// nope, no module
	if module.ID == uuid.Nil {
		c.Status(http.StatusNotFound)

		view.RenderErrTemplate(c, "module/404.tmpl", nil)
		return nil, nil
	}

	return
}

// BuildBook runs mdbook on the /doc (or otherwise configured) folder of the module to generate documentation
func BuildBook(c *gin.Context) {
	_, module := getModule(c)

	// getModule already writes out the necessary error messages
	if module == nil {
		return
	}

	// get the repo cache
	g := &repo.Git{URL: module.RepoURL}
	var docPath string
	var err error

	if module.Sub != "" {
		docPath, err = g.SubModuleDocs(module.Sub)
	}

	repoPath, err := g.Dir()

	repoDocPath := filepath.Join(repoPath, docPath)

	ctx, cancel := context.WithTimeout(c, 60*time.Second)
	defer cancel()

	dest := filepath.Join(config.Cfg.Cache.Book.Base, module.ID.String())

	zap.L().Debug("book destination", zap.String("path", dest))

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err := os.Mkdir(dest, 0755)
		if err != nil {
			zap.L().Panic("could not create book dir", zap.Error(err))
		}
	}

	// build the html pages with mdbook
	argv := []string{"build", repoDocPath, "-d", dest}
	bookCmd := exec.CommandContext(ctx, "mdbook", argv...)
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
		view.RenderTemplate(c, "bench/merge_error.tmpl", "mdbook Error", m)
		return
	}

	zap.L().Debug("build book log output", zap.ByteString("output", logOutput))

	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/module/doc/%s", module.ID))
}

func PullAllRepos(c *gin.Context) {
	var modules []model.Module

	result := model.DB.Preload("User").Preload("Category").Find(&modules)
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

		mod.Metadata = meta

		result = model.DB.WithContext(c).Save(mod)
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
