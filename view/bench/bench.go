package bench

// SPDX-License-Identifier: EUPL-1.2

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/merge"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
	"gorm.io/gorm"
)

func viewHelper(id, tmpl string, w http.ResponseWriter, r *http.Request) {
	// View the current active bench or another bench if a parameter is supplied
	bench := &model.Bench{}
	ctx := r.Context()
	user := ctx.Value(util.UserContextKey).(*model.User)

	// check if we even have a module id
	if id == "" {
		if user != nil {
			result := model.DB.Model(bench).Where("user_id = ? and active = true", user.ID).First(bench)
			if result.Error != nil {
				view.RenderErrMarkdown(ctx, w, "bench/404.md", result.Error)
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
	result := model.DB.Where("id = ? and (public = true or user_id = ?)", id, user.ID).First(bench)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Panic().Err(result.Error).Msgf("could not get the bench")
	}

	// nope, no bench
	if bench.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		view.RenderMarkdown("bench/404.md", nil, w)
		return
	}

	// load further information
	var benchMods []model.BenchModule
	result = model.DB.Preload("Module").Where("bench_id = ?", id).Find(&benchMods)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not get the bench modules")
	}

	// get bench macro parameters (future)

	// all packed up,
	m := map[string]interface{}{
		"Bench":   bench,
		"User":    user,
		"Modules": benchMods,
		"Author":  mup.DisplayName,
		"Title":   fmt.Sprintf("EDeA - %s", bench.Name),
		"Error":   nil,
	}

	// and ready to go
	view.RenderTemplate(ctx, tmpl, "", m, w)
}

// View a Bench
func View(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	viewHelper(id, "bench/view.tmpl", w, r)
}

// ViewUpdate is the same as view but renders the update form for a bench
func ViewUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	viewHelper(id, "bench/update.tmpl", w, r)
}

// SetActive sets the requested bench as active and inactivates all the others
func SetActive(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	benchID := vars["id"]
	user := view.CurrentUser(r)

	// TODO: more error checking and input validation

	// set other benches as inactive, activate the requested one
	tx := model.DB.WithContext(r.Context()).Begin()

	tx.Model(&model.Bench{}).Where("user_id = ? and active = true", user.ID).Update("active", false)
	tx.Model(&model.Bench{}).Where("id = ? and user_id = ?", benchID, user.ID).Update("active", true)

	if tx.Error != nil {
		log.Panic().Err(tx.Error).Msg("could not update benches")
		tx.Rollback()
	} else {
		tx.Commit()
	}

	// redirect to the bench
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", benchID), http.StatusSeeOther)
}

// Delete a bench
func Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	benchID := vars["id"]
	ctx := r.Context()
	user := ctx.Value(util.UserContextKey).(*model.User)

	if benchID == "" {
		view.RenderErrTemplate(ctx, w, "404.tmpl", fmt.Errorf(`no such bench: "%s"`, benchID))
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	result := model.DB.Where("user_id = ?", user.ID).Delete(&model.Bench{}, uuid.MustParse(benchID))
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not delete bench")
	}

	http.Redirect(w, r, "/bench/user/me", http.StatusTemporaryRedirect)
}

// Fork a bench, this only copies it to the current user as we don't have any versioning (yet)
func Fork(w http.ResponseWriter, r *http.Request) {
	// Fork a bench
	ctx := r.Context()
	user := ctx.Value(util.UserContextKey).(*model.User)

	vars := mux.Vars(r)
	id := vars["id"]

	b := &model.Bench{}

	result := model.DB.Model(b).Where("id = ? and (user_id = ? OR public = true)", id, user.ID).First(b)
	if result.Error != nil {
		view.RenderErrMarkdown(ctx, w, "bench/404.md", result.Error)
		return
	}

	// no such (public) bench
	if b.ID == uuid.Nil {
		view.RenderErrTemplate(ctx, w, "bench/404.tmpl", fmt.Errorf("could not find bench or bench is private"))
		return
	}

	// load all the modules + configuration as we need to clone them too
	var benchMods []model.BenchModule
	result = model.DB.Preload("Module").Where("bench_id = ?", id).Find(&benchMods)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not get the bench modules")
	}

	// create new bench here
	b.ID = uuid.Nil
	b.UserID = user.ID
	b.Public = false
	b.Active = true

	tx := model.DB.WithContext(r.Context()).Begin()

	tx.Model(&model.Bench{}).Where("user_id = ? and active = true", user.ID).Update("active", false)
	tx.Create(b)

	if tx.Error != nil {
		tx.Rollback()
		log.Panic().Err(tx.Error).Msg("could not create the new benche")
	} else {
		tx.Commit()
	}

	var err error

	tx = model.DB.Begin()
	for _, m := range benchMods {
		m.ID = uuid.Nil
		m.BenchID = b.ID
		if result := tx.Create(&m); result.Error != nil {
			err = result.Error
			break
		}
	}

	if err == nil {
		err = tx.Commit().Error
	}

	if err != nil {
		log.Error().Err(err).Msgf("could not fork bench %s", id)
		tx.Rollback()

		// try to remove the bench we already created now
		if result := model.DB.Delete(b); result.Error != nil {
			log.Panic().Err(err).Msgf("could not remove newly created bench %s", b.ID.String())
		}

		view.RenderErrTemplate(ctx, w, "500.tmpl", fmt.Errorf("could not fork the bench"))
		return
	}

	// if everything went well, present the user with a newly forked bench
	viewHelper(b.ID.String(), "bench/update.tmpl", w, r)
}

// Create inserts a new bench
func Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value(util.UserContextKey).(*model.User)

	if err := r.ParseForm(); err != nil {
		view.RenderErrMarkdown(ctx, w, "bench/new.md", err)
		return
	}

	bench := new(model.Bench)
	if err := util.FormDecoder.Decode(bench, r.Form); err != nil {
		view.RenderErrMarkdown(ctx, w, "bench/new.md", err)
		return
	}

	bench.ID = uuid.Nil // prevent the client setting an id
	bench.Active = true
	bench.UserID = user.ID

	// set other benches as inactive, activate the requested one
	tx := model.DB.WithContext(ctx).Begin()

	tx.Model(&model.Bench{}).Where("user_id = ? and active = true", user.ID).Update("active", false)
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
		view.RenderErrMarkdown(r.Context(), w, "bench/update.md", err)
		return
	}

	bench := new(model.Bench)
	if err := util.FormDecoder.Decode(bench, r.Form); err != nil {
		view.RenderErrMarkdown(r.Context(), w, "bench/update.md", err)
		return
	}

	log.Debug().Msgf("%+v", bench)

	// make sure that we update only the fields a user should be able to change
	result := model.DB.Model(bench).Select("Name", "Description", "Public").Updates(bench)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not update bench")
	}

	// redirect to the bench
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}

// Current redirects to the users active bench
func Current(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value(util.UserContextKey).(*model.User)

	bench := new(model.Bench)

	result := model.DB.WithContext(ctx).Where("user_id = ? and active = true", user.ID).Find(bench)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not create a new bench")
	}

	if bench.ID == uuid.Nil {
		http.Redirect(w, r, "/bench/user/me", http.StatusSeeOther)
		return
	}

	// redirect to newly created module page
	http.Redirect(w, r, fmt.Sprintf("/bench/%s", bench.ID), http.StatusSeeOther)
}

// ListUser lists all Benches belonging to a User
func ListUser(w http.ResponseWriter, r *http.Request) {
	var result *gorm.DB
	var benches []model.Bench
	var user *model.User

	vars := mux.Vars(r)
	userID := vars["id"]
	ctx := r.Context()

	if v := ctx.Value(util.UserContextKey); v != nil {
		user = v.(*model.User)
	}

	m := make(map[string]interface{})

	if userID == "me" {
		if user != nil {
			// select a users own benches
			result = model.DB.Where("user_id = ?", user.ID, userID).Find(&benches)
		} else {
			// a user must be logged in to see their own benches
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	} else {
		// list another users public benches
		uid := uuid.MustParse(userID)
		mup := model.Profile{UserID: uid}

		if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
			log.Error().Err(result.Error).Msgf("could not fetch bench author profile for user_id %s", uid)
		}

		m["Author"] = mup

		result = model.DB.Where("user_id = ? and public = true", user.ID, userID).Find(&benches)
	}

	if result.Error != nil {
		view.RenderErrMarkdown(ctx, w, "bench/404.md", result.Error)
		return
	}

	// all packed up,
	m["Benches"] = benches
	m["Error"] = nil

	// and ready to go
	view.RenderTemplate(ctx, "bench/list_user.tmpl", "", m, w)
}

// Merge a bench into a new kicad project
func Merge(w http.ResponseWriter, r *http.Request) {
	var userID uuid.UUID

	vars := mux.Vars(r)
	id := vars["id"]
	ctx := r.Context()

	bench := new(model.Bench)

	user, ok := ctx.Value(util.UserContextKey).(*model.User)
	if ok {
		userID = user.ID
	}

	// try to fetch all the benchmodules
	result := model.DB.WithContext(ctx).Preload("Modules.Module").Where("id = ? AND (user_id = ? OR public = true)", id, userID).Find(bench)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not create a new bench")
	}

	if bench.ID == uuid.Nil {
		log.Panic().Msg("could not find bench")
	}

	// and merge it together
	b, err := merge.Merge(bench.Name, bench.Modules)
	if err != nil {
		log.Panic().Err(err).Msg("something went wrong during merge")
	}

	buf := bytes.NewReader(b)

	fileName := fmt.Sprintf("%s.zip", bench.Name)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	http.ServeContent(w, r, fileName, time.Now(), buf)
}
