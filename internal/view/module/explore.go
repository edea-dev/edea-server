package module

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/edea-dev/edead/internal/model"
	"gitlab.com/edea-dev/edead/internal/util"
	"gitlab.com/edea-dev/edead/internal/view"
	"go.uber.org/zap"
)

// TOOD: improve DDL so we can do this with the ORM natively

type ExploreModule struct {
	ID          string
	UserID      string
	DisplayName string
	Name        string
	Description string
	RepoURL     string
	Category    string
}

const exploreQuery = `
	SELECT m.id, m.user_id, p.display_name, m.name, m.description, m.repo_url, c.name as category
	FROM modules m
	JOIN profiles p
		ON p.user_id = m.user_id
	JOIN categories c
		ON c.id = m.category_id
	WHERE m.private = false
		AND m.deleted_at IS NULL
	ORDER BY m.updated_at;`

// Explore modules page
func Explore(c *gin.Context) {
	var p []ExploreModule

	result := model.DB.Raw(exploreQuery).Scan(&p)
	if result.Error != nil {
		zap.L().Panic("could not run explore query", zap.Error(result.Error))
	}

	m := map[string]interface{}{
		"Modules": p,
	}

	view.RenderTemplate(c, "explore/view.tmpl", "EDeA - Explore Modules", m)
}

// ExploreUser lists all the public modules by a specific user
func ExploreUser(c *gin.Context) {
	userID := c.Param("id")
	currentUser, _ := c.Value(util.UserContextKey).(*model.User)

	if userID == "" {
		zap.L().Panic("explore_user: no user id specified")
	}

	var id uuid.UUID

	// /explore/user/me should return your own modules
	if userID == "me" {
		// redirect to the login page if someone isn't logged in
		if currentUser == nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}
		id = currentUser.ID
	} else {
		id = uuid.MustParse(userID)
	}

	p := &model.Profile{UserID: id}

	result := model.DB.Find(&p)
	if result.Error != nil {
		zap.L().Panic("no such user", zap.Error(result.Error))
	}

	var modules []model.Module
	m := make(map[string]interface{})

	if currentUser != nil && id == currentUser.ID {
		// show a users own modules
		result = model.DB.Where(&model.Module{UserID: id}).Find(&modules)
		m["Self"] = true
	} else {
		// show another users modules
		result = model.DB.Where("user_id = ? AND private = false AND deleted_at IS NULL", id).Find(&modules)
	}
	if result.Error != nil {
		zap.L().Panic("could not run user explore query", zap.Error(result.Error))
	}

	m["Modules"] = modules
	m["Profile"] = p

	view.RenderTemplate(c, "explore/user.tmpl", "EDeA - User Modules", m)
}
