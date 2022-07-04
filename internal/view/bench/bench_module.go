package bench

// SPDX-License-Identifier: EUPL-1.2

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/edea-dev/edea-server/internal/model"
	"gitlab.com/edea-dev/edea-server/internal/util"
	"gitlab.com/edea-dev/edea-server/internal/view"
	"go.uber.org/zap"
)

// AddModule adds a module to the currently active bench
func AddModule(c *gin.Context) {
	moduleID := c.Param("id")
	user := view.CurrentUser(c)

	// we need to actually have an id to add it to the current bench
	if moduleID == "" {
		view.RenderErrTemplate(c, "module/add_err.md", util.ErrImSorryDave)
		return
	}

	module := new(model.Module)

	// get the module by id but also check if it belongs to the user requesting it in case its a private module
	result := model.DB.Where("id = ? and (user_id = ? or private = false)", moduleID, user.ID).Find(module)
	if result.Error != nil {
		view.RenderErrTemplate(c, "module/add_err.md", util.ErrNoSuchModule)
		return
	}

	// get the currently active bench
	bench := new(model.Bench)
	result = model.DB.Where("user_id = ? and active = true", user.ID).Find(bench)
	if result.Error != nil {
		view.RenderErrTemplate(c, "module/add_err.md", util.ErrNoSuchBench)
		return
	}

	// create a new bench in this case for convenience. a user should *usually* have a bench active
	if bench.ID == uuid.Nil {
		bench.Active = true
		bench.Name = "New Bench"
		bench.UserID = user.ID

		// set other benches as inactive, activate the requested one
		result := model.DB.WithContext(c).Create(bench)

		if result.Error != nil {
			zap.L().Panic("could not create a new bench", zap.Error(result.Error))
		}

		zap.L().Debug("new bench", zap.Object("bench", bench))
	}

	benchModule := &model.BenchModule{BenchID: bench.ID, ModuleID: module.ID, Name: module.Name}

	result = model.DB.Create(benchModule)
	if result.Error != nil {
		zap.L().Panic("could not create a new bench_module for bench", zap.Error(result.Error), zap.String("bench_id", bench.ID.String()))
	}

	// redirect to newly created bench page
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/bench/%s", bench.ID))
}

// RemoveModule removes a module from the currently active bench
func RemoveModule(c *gin.Context) {
	benchModuleID := c.Param("id")
	user := view.CurrentUser(c)

	// we need to actually have an id to add it to the current bench
	if benchModuleID == "" {
		view.RenderErrTemplate(c, "module/remove_err.md", util.ErrImSorryDave)
		return
	}

	// get the currently active bench
	bench := new(model.Bench)
	result := model.DB.Where("user_id = ? and active = true", user.ID).Find(bench)
	if result.Error != nil {
		view.RenderErrTemplate(c, "module/remove_err.md", util.ErrNoSuchBench)
		return
	}

	if bench.ID == uuid.Nil {
		// user tried to remove a module from an inactive bench, this will be supported through the API only
		view.RenderErrTemplate(c, "module/remove_err.md", util.ErrNoActiveBench)
		return
	}

	result = model.DB.Where("id = ?", benchModuleID).Delete(&model.BenchModule{})
	if result.Error != nil {
		zap.L().Panic("could not remove bench_module from bench", zap.Error(result.Error), zap.String("bench_id", bench.ID.String()))
	}

	// redirect to the current bench
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/bench/%s", bench.ID))
}
