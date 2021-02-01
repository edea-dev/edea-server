package bench

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
)

// AddModule adds a module to the currently active bench
func AddModule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	moduleID := vars["id"]
	user := view.CurrentUser(r)

	// we need to actually have an id to add it to the current bench
	if moduleID == "" {
		view.RenderErr(r.Context(), w, "module/add_err.md", util.ErrImSorryDave)
		return
	}

	module := new(model.Module)

	// get the module by id but also check if it belongs to the user requesting it in case its a private module
	result := model.DB.Where("id = ? and (user_id = ? or private = false)", moduleID, user.ID).Find(module)
	if result.Error != nil {
		view.RenderErr(r.Context(), w, "module/add_err.md", util.ErrNoSuchModule)
		return
	}

	// get the currently active bench
	bench := new(model.Bench)
	result = model.DB.Where("user_id = ? and active = true", user.ID).Find(bench)
	if result.Error != nil {
		view.RenderErr(r.Context(), w, "module/add_err.md", util.ErrNoSuchBench)
		return
	}

	// TODO: create a new bench in this case for convenience. a user should *usually* have a bench active
	if bench.ID == uuid.Nil {
		view.RenderErr(r.Context(), w, "module/add_err.md", util.ErrNoActiveBench)
		return
	}

	benchModule := &model.BenchModule{BenchID: bench.ID, ModuleID: module.ID, Name: module.Name}

	result = model.DB.Create(benchModule)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not add a new bench module to %s", bench.ID)
	}

	// redirect to newly created module page
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}
