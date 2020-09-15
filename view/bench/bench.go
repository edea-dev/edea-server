package bench

import (
	"fmt"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
)

// View a Bench
func View(w http.ResponseWriter, r *http.Request) {
	// View the current active bench or another bench if a parameter is supplied
	vars := mux.Vars(r)
	benchID := vars["id"]
	bench := &model.Bench{}
	user := view.CurrentUser(r)

	// check if we even have a module id
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
			view.RenderMarkdown("bench/404.md", msg, w)
			return
		}
	}

	// try to fetch the bench, TODO: join with modules
	err := model.DB.Model(bench).Where("id = ? and (private = false or user_id = ?)", benchID, user.ID).Select()
	if err != nil {
		log.Panic().Err(err).Msgf("could not get the bench: %v", err)
	}

	// nope, no bench
	if bench.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("bench/404.md", nil, w)
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

// SetActive sets the requested bench as active and inactivates all the others
func SetActive(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "bench/new.md", err)
		return
	}

	bench := new(model.Bench)
	if err := util.FormDecoder.Decode(bench, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "bench/new.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)
	userID := user.ID

	if bench.UserID != user.ID {
		if user.IsAdmin {
			// admin changing stuff
			userID = bench.UserID
		} else {
			view.RenderErr(r.Context(), w, "bench/new.md", util.ErrImSorryDave)
			return
		}
	}

	// set other benches as inactive, activate the requested one
	err := model.DB.RunInTransaction(r.Context(), func(tx *pg.Tx) error {
		_, err := tx.Exec("update bench set active = false where user_id = ?", userID)
		if err != nil {
			return err
		}

		res, err := tx.Exec("update bench set active = true where bench_id = ? and user_id = ?", bench.ID, userID)
		if err != nil {
			return err
		}
		if res.RowsAffected() != 1 {
			return util.ErrNoSuchBench
		}
		return nil
	})

	if err != nil {
		log.Panic().Err(err).Msg("could not update benches")
	}

	// redirect to the bench
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	// Delete a bench
}

func Fork(w http.ResponseWriter, r *http.Request) {
	// Fork a bench
}

// Create inserts a new bench
func Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "bench/new.md", err)
		return
	}

	bench := new(model.Bench)
	if err := util.FormDecoder.Decode(bench, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "bench/new.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	if bench.UserID != user.ID {
		if user.IsAdmin {
			// TODO: admin changing stuff
		} else {
			view.RenderErr(r.Context(), w, "bench/new.md", util.ErrImSorryDave)
			return
		}
	}

	bench.ID = "" // prevent the client setting an id
	bench.Active = true

	// set other benches as inactive
	err := model.DB.RunInTransaction(r.Context(), func(tx *pg.Tx) error {
		_, err := tx.Exec("update bench set active = false where user_id = ?", user.ID)
		if err != nil {
			return err
		}

		_, err = tx.Model(bench).Insert()
		return err
	})

	if err != nil {
		log.Panic().Err(err).Msg("could not create a new bench")
	}

	// redirect to newly created module page
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}

// Update a bench
func Update(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "bench/update.md", err)
		return
	}

	bench := new(model.Bench)
	if err := util.FormDecoder.Decode(bench, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "bench/update.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	if bench.UserID != user.ID {
		if user.IsAdmin {
			// TODO: admin is changing stuff
		} else {
			view.RenderErr(r.Context(), w, "bench/update.md", util.ErrImSorryDave)
			return
		}
	}

	// set other benches as inactive, activate the requested one
	_, err := model.DB.Model(bench).Update()
	if err != nil {
		log.Panic().Err(err).Msg("could not update bench")
	}

	// redirect to the bench
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}

// New Bench view
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
