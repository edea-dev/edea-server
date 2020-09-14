package bench

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/view"
)

// View a Bench
func View(w http.ResponseWriter, r *http.Request) {
	// View the current active bench or another bench if a parameter is supplied
	vars := mux.Vars(r)
	benchID := vars["id"]
	bench := &model.Bench{}
	user := view.CurrentUser(r)

	// check if we even have a project id
	if benchID == "" {
		if user != nil {
			err := model.DB.Model(bench).Where("user_id = ? and active = true", user.ID).Select()
			if err != nil {
				view.RenderErr(r.Context(), w, "bench/404.md", err)
				return
			}
		} else {
			msg := map[string]interface{}{
				"Error": "Unfortunately you didn't give us much to work with, try again with a bench id.",
			}
			w.WriteHeader(http.StatusNotFound)
			view.RenderMarkdown("project/404.md", msg, w)
			return
		}
	}

	// try to fetch the bench, TODO: join with projects
	err := model.DB.Model(bench).Where("id = ? and (private = false or user_id = ?)", benchID, user.ID).Select()
	if err != nil {
		log.Panic().Err(err).Msgf("could not get the project: %v", err)
	}

	// nope, no project
	if bench.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("project/404.md", nil, w)
		return
	}

	// load further information?

	// get bench macro parameters (future)

	// all packed up,
	m := map[string]interface{}{
		"Bench": bench,
		"User":  user,
		"Error": err,
	}

	// and ready to go
	view.RenderMarkdown("bench/view.md", m, w)
}

func SetActive(w http.ResponseWriter, r *http.Request) {
	// Set another bench active for the current user
	// set all benches to inactive for user, set this one active
}

func Delete(w http.ResponseWriter, r *http.Request) {
	// Delete a bench
}

func Fork(w http.ResponseWriter, r *http.Request) {
	// Fork a bench
}
