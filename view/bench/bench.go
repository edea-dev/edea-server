package bench

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
	"gorm.io/gorm"
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
			result := model.DB.Model(bench).Where("user_id = ? and active = true", user.ID).First(bench)
			if result.Error != nil {
				view.RenderErr(r.Context(), w, "bench/404.md", result.Error)
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

	// get the bench author name
	mup := model.Profile{UserID: bench.UserID}

	if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
		log.Error().Err(result.Error).Msgf("could not fetch bench author profile for user_id %s", bench.UserID)
	}

	// try to fetch the bench, TODO: join with modules
	result := model.DB.Where("id = ? and (public = true or user_id = ?)", benchID, user.ID).First(bench)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not get the bench")
	}

	// nope, no bench
	if bench.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("bench/404.md", nil, w)
		return
	}

	// load further information?

	// get bench macro parameters (future)

	// all packed up,
	m := map[string]interface{}{
		"Bench":  bench,
		"User":   user,
		"Author": mup.DisplayName,
		"Error":  nil,
	}

	// and ready to go
	view.RenderMarkdown("bench/view.md", m, w)
}

// SetActive sets the requested bench as active and inactivates all the others
func SetActive(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "bench/view.md", err)
		return
	}

	bench := new(model.Bench)
	if err := util.FormDecoder.Decode(bench, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "bench/view.md", err)
		return
	}

	// set other benches as inactive, activate the requested one
	tx := model.DB.WithContext(r.Context()).Begin()

	tx.Model(bench).Where("user_id = ? and active = true", bench.User.ID).Update("active", false)
	tx.Where(bench).Update("active", true)

	if tx.Error != nil {
		log.Panic().Err(tx.Error).Msg("could not update benches")
		tx.Rollback()
	} else {
		tx.Commit()
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
	user := r.Context().Value(util.UserContextKey).(*model.User)

	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "bench/new.md", err)
		return
	}

	bench := new(model.Bench)
	if err := util.FormDecoder.Decode(bench, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "bench/new.md", err)
		return
	}

	bench.ID = uuid.Nil // prevent the client setting an id
	bench.Active = true
	bench.UserID = user.ID

	// set other benches as inactive, activate the requested one
	tx := model.DB.WithContext(r.Context()).Begin()

	tx.Model(bench).Where("user_id = ? and active = true", bench.User.ID).Update("active", false)
	tx.Create(bench)
	tx.Commit()

	if tx.Error != nil {
		log.Panic().Err(tx.Error).Msg("could not create a new bench")
		tx.Rollback()
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

	// set other benches as inactive, activate the requested one
	result := model.DB.Save(bench)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not update bench")
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

	view.RenderMarkdown("bench/new.md", data, w)
}

// Current redirects to the users active bench
func Current(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.UserContextKey).(*model.User)

	bench := new(model.Bench)

	result := model.DB.WithContext(r.Context()).Where("user_id = ? and active = true", user.ID).Find(bench)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Redirect(w, r, "/bench/list", http.StatusSeeOther)
		} else {
			log.Panic().Err(result.Error).Msg("could not create a new bench")
		}
	}

	// redirect to newly created module page
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}

// ListUser lists all Benches belonging to a User
func ListUser(w http.ResponseWriter, r *http.Request) {
	user := view.CurrentUser(r)
	var benches []model.Bench

	// check if we even have a module id

	if user != nil {
		result := model.DB.Where("user_id = ?", user.ID).Find(&benches)
		if result.Error != nil {
			view.RenderErr(r.Context(), w, "bench/404.md", result.Error)
			return
		}
	} else {
		// a user must be logged in to see their own benches
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// all packed up,
	m := map[string]interface{}{
		"Benches": benches,
		"User":    user,
		"Error":   nil,
	}

	// and ready to go
	view.RenderMarkdown("bench/list_user.md", m, w)
}
