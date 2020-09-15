package bench

import (
	"fmt"
	"net/http"

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
	err := model.DB.Model(module).
		Where("id = ? and (user_id = ? or private = false)", module.ID, user.ID).
		Select()

	// get the currently active bench
	bench := new(model.Bench)
	err = model.DB.Model(bench).Where("user_id = ? and active = true", user.ID).Select()
	if err != nil {
		view.RenderErr(r.Context(), w, "module/add_err.md", util.ErrNoSuchBench)
		return
	}

	// TODO: create a new bench in this case for convenience. a user should *usually* have a bench active
	if bench.ID == "" {
		view.RenderErr(r.Context(), w, "module/add_err.md", util.ErrNoActiveBench)
		return
	}

	benchModule := &model.BenchModule{BenchID: bench.ID, ModuleID: module.ID, Name: module.Name}

	_, err = model.DB.Model(benchModule).Insert()
	if err != nil {
		log.Panic().Err(err).Msgf("could not add a new bench module to %s", bench.ID)
	}

	// redirect to newly created module page
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}
