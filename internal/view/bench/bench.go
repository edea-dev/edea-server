package bench

// SPDX-License-Identifier: EUPL-1.2

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/edea-dev/edead/internal/merge"
	"gitlab.com/edea-dev/edead/internal/model"
	"gitlab.com/edea-dev/edead/internal/search"
	"gitlab.com/edea-dev/edead/internal/util"
	"gitlab.com/edea-dev/edead/internal/view"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func viewHelper(id, tmpl string, c *gin.Context) {
	// View the current active bench or another bench if a parameter is supplied
	bench := &model.Bench{}
	user, _ := c.Keys["user"].(*model.User)

	// check if we even have a module id
	if id == "" {
		if user != nil {
			result := model.DB.Model(bench).Where("user_id = ? and active = true", user.ID).First(bench)
			if result.Error != nil {
				view.RenderErrMarkdown(c, "bench/404.md", result.Error)
				return
			}
		} else {
			msg := map[string]interface{}{
				"Error": "Unfortunately you didn't give us much to work with, try again with a bench id.",
			}
			c.Status(http.StatusNotFound)
			view.RenderMarkdown(c, "bench/404.md", msg)
			return
		}
	}

	// get the bench author name
	mup := model.Profile{UserID: bench.UserID}

	if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
		zap.L().Error("could not fetch bench author profile", zap.Error(result.Error), zap.String("bench_user_id", bench.UserID.String()))
	}

	// try to fetch the bench, TODO: join with modules
	var result *gorm.DB
	if user == nil {
		result = model.DB.Where("id = ? and (public = true)", id).First(bench)
	} else {
		result = model.DB.Where("id = ? and (public = true or user_id = ?)", id, user.ID).First(bench)
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		zap.L().Panic("could not get the bench", zap.Error(result.Error))
	}

	// nope, no bench
	if bench.ID == uuid.Nil {
		c.Status(http.StatusNotFound)
		view.RenderErrTemplate(c, "bench/404.tmpl", errors.New("Bench was not found or is private"))
		return
	}

	// load further information
	var benchMods []model.BenchModule
	result = model.DB.Preload("Module").Where("bench_id = ?", id).Find(&benchMods)
	if result.Error != nil {
		zap.L().Panic("could not get the bench modules", zap.Error(result.Error))
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
	view.RenderTemplate(c, tmpl, "", m)
}

// View a Bench
func View(c *gin.Context) {
	viewHelper(c.Param("id"), "bench/view.tmpl", c)
}

// ViewUpdate is the same as view but renders the update form for a bench
func ViewUpdate(c *gin.Context) {
	viewHelper(c.Param("id"), "bench/update.tmpl", c)
}

// SetActive sets the requested bench as active and inactivates all the others
func SetActive(c *gin.Context) {
	benchID := c.Param("id")
	user := view.CurrentUser(c)

	// TODO: more error checking and input validation

	// set other benches as inactive, activate the requested one
	tx := model.DB.WithContext(c).Begin()

	tx.Model(&model.Bench{}).Where("user_id = ? and active = true", user.ID).Update("active", false)
	tx.Model(&model.Bench{}).Where("id = ? and user_id = ?", benchID, user.ID).Update("active", true)

	if tx.Error != nil {
		zap.L().Panic("could not update benches", zap.Error(tx.Error))
		tx.Rollback()
	} else {
		tx.Commit()
	}

	// redirect to the bench
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/bench/%s", benchID))
}

// Delete a bench
func Delete(c *gin.Context) {
	benchID := c.Param("id")
	user := c.Keys["user"].(*model.User)

	if benchID == "" {
		view.RenderErrTemplate(c, "404.tmpl", fmt.Errorf(`no such bench: "%s"`, benchID))
		return
	}

	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	result := model.DB.Where("user_id = ?", user.ID).Delete(&model.Bench{}, uuid.MustParse(benchID))
	if result.Error != nil {
		zap.L().Panic("could not delete bench", zap.Error(result.Error))
	}

	// update search index
	if err := search.DeleteEntry(search.Entry{ID: benchID}); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	c.Redirect(http.StatusTemporaryRedirect, "/bench/user/me")
}

// Fork a bench, this only copies it to the current user as we don't have any versioning (yet)
func Fork(c *gin.Context) {
	// Fork a bench
	user := c.Keys["user"].(*model.User)
	id := c.Param("id")

	b := &model.Bench{}

	result := model.DB.Model(b).Where("id = ? and (user_id = ? OR public = true)", id, user.ID).First(b)
	if result.Error != nil {
		view.RenderErrMarkdown(c, "bench/404.md", result.Error)
		return
	}

	// no such (public) bench
	if b.ID == uuid.Nil {
		view.RenderErrTemplate(c, "bench/404.tmpl", fmt.Errorf("could not find bench or bench is private"))
		return
	}

	// load all the modules + configuration as we need to clone them too
	var benchMods []model.BenchModule
	result = model.DB.Preload("Module").Where("bench_id = ?", id).Find(&benchMods)
	if result.Error != nil {
		zap.L().Panic("could not get the bench modules", zap.Error(result.Error))
	}

	// create new bench here
	b.ID = uuid.Nil
	b.UserID = user.ID
	b.Public = false
	b.Active = true

	tx := model.DB.WithContext(c).Begin()

	tx.Model(&model.Bench{}).Where("user_id = ? and active = true", user.ID).Update("active", false)
	tx.Create(b)

	if tx.Error != nil {
		tx.Rollback()
		zap.L().Panic("could not create the new bench", zap.Error(tx.Error))
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
		zap.L().Error("could not fork bench", zap.Error(result.Error), zap.String("bench_id", id))
		tx.Rollback()

		// try to remove the bench we already created now
		if result := model.DB.Delete(b); result.Error != nil {
			zap.L().Panic("could not remove newly created bench", zap.Error(result.Error), zap.String("bench_id", b.ID.String()))
		}

		view.RenderErrTemplate(c, "500.tmpl", fmt.Errorf("could not fork the bench"))
		return
	}

	// update search index
	if err := search.UpdateEntry(search.BenchToEntry(*b)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	// if everything went well, present the user with a newly forked bench
	viewHelper(b.ID.String(), "bench/update.tmpl", c)
}

// Create inserts a new bench
func Create(c *gin.Context) {
	user := c.Keys["user"].(*model.User)

	bench := new(model.Bench)
	if err := c.Bind(bench); err != nil {
		view.RenderErrMarkdown(c, "bench/new.md", err)
		return
	}

	bench.ID = uuid.Nil // prevent the client setting an id
	bench.Active = true
	bench.UserID = user.ID

	// set other benches as inactive, activate the requested one
	tx := model.DB.WithContext(c).Begin()

	tx.Model(&model.Bench{}).Where("user_id = ? and active = true", user.ID).Update("active", false)
	tx.Create(bench)
	tx.Commit()

	if tx.Error != nil {
		zap.L().Panic("could not create a new bench", zap.Error(tx.Error))

		tx.Rollback()
	}

	// update search index
	if err := search.UpdateEntry(search.BenchToEntry(*bench)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	// redirect to newly created module page
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/bench/%s", bench.ID))
}

// Update a bench
func Update(c *gin.Context) {
	bench := &model.Bench{}

	if err := c.Bind(bench); err != nil {
		view.RenderErrMarkdown(c, "bench/update.md", err)
		return
	}

	// log.Debug().Msgf("%+v", bench)

	// make sure that we update only the fields a user should be able to change
	result := model.DB.Model(bench).Select("Name", "Description", "Public").Updates(bench)
	if result.Error != nil {
		zap.L().Panic("could not update bench", zap.Error(result.Error))
	}

	// update search index
	if err := search.UpdateEntry(search.BenchToEntry(*bench)); err != nil {
		zap.L().Panic("could not update search index", zap.Error(err))
	}

	// redirect to the bench
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/bench/%s", bench.ID))
}

// Current redirects to the users active bench
func Current(c *gin.Context) {
	user := c.Keys["user"].(*model.User)

	bench := new(model.Bench)

	result := model.DB.WithContext(c).Where("user_id = ? and active = true", user.ID).Find(bench)
	if result.Error != nil {
		zap.L().Panic("could not create a new bench", zap.Error(result.Error))
	}

	if bench.ID == uuid.Nil {
		c.Redirect(http.StatusSeeOther, "/bench/user/me")
		return
	}

	// redirect to newly created module page
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/bench/%s", bench.ID))
}

// ListUser lists all Benches belonging to a User
func ListUser(c *gin.Context) {
	var result *gorm.DB
	var benches []model.Bench
	var user *model.User

	userID := c.Param("id")

	if v := c.Keys["user"]; v != nil {
		user = v.(*model.User)
	}

	m := make(map[string]interface{})

	if userID == "me" {
		if user != nil {
			// select a users own benches
			result = model.DB.Where("user_id = ?", user.ID).Find(&benches)
		} else {
			// a user must be logged in to see their own benches
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
	} else {
		// list another users public benches
		uid := uuid.MustParse(userID)
		mup := model.Profile{UserID: uid}

		if result := model.DB.Where(&mup).First(&mup); result.Error != nil {
			zap.L().Error("could not fetch bench author profile", zap.Error(result.Error), zap.String("user_id", uid.String()))
		}

		m["Author"] = mup

		result = model.DB.Where("user_id = ? and public = true", userID).Find(&benches)
	}

	if result.Error != nil {
		view.RenderErrMarkdown(c, "bench/404.md", result.Error)
		return
	}

	// all packed up,
	m["Benches"] = benches
	m["Error"] = nil

	// and ready to go
	view.RenderTemplate(c, "bench/list_user.tmpl", "", m)
}

// Merge a bench into a new kicad project
func Merge(c *gin.Context) {
	var userID uuid.UUID

	id := c.Param("id")

	bench := new(model.Bench)

	user, ok := c.Keys["user"].(*model.User)
	if ok {
		userID = user.ID
	}

	// try to fetch all the benchmodules
	result := model.DB.WithContext(c).Preload("Modules.Module").Where("id = ? AND (user_id = ? OR public = true)", id, userID).Find(bench)
	if result.Error != nil {
		zap.L().Panic("could not create a new bench", zap.Error(result.Error))
	}

	if bench.ID == uuid.Nil {
		zap.L().Panic("could not find bench", zap.Error(result.Error))
	}

	// and merge it together
	b, err := merge.Merge(bench.Name, bench.Modules)

	// show the user the tool output in case of an error while merging
	if err != nil {
		m := map[string]interface{}{
			"Error":  err,
			"Output": strings.ReplaceAll(string(b), "\n", "<br>"),
		}
		if err, ok := err.(util.HintError); ok {
			zap.L().Debug("error with hint", zap.Error(err.Err), zap.String("hint", err.Hint))
			m["Error"] = err.Err
			m["Hint"] = err.Hint
		}
		view.RenderTemplate(c, "bench/merge_error.tmpl", "Merge Error", m)
		return
	}

	buf := bytes.NewReader(b)

	fileName := fmt.Sprintf("%s.zip", bench.Name)

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	http.ServeContent(c.Writer, c.Request, fileName, time.Now(), buf)
}
