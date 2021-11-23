package bench

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/model"
	"gitlab.com/edea-dev/edead/internal/view"
	"go.uber.org/zap"
)

// ExploreBench struct
type ExploreBench struct {
	ID          string
	UserID      string
	DisplayName string
	Name        string
	Description string
}

const exploreQuery = `
	SELECT b.id, b.user_id, p.display_name, b.name, b.description
	FROM benches b
	JOIN profiles p
		ON p.user_id = b.user_id
	WHERE b.public = true
		AND b.deleted_at IS NULL
	ORDER BY b.updated_at;`

const exploreOwnQuery = `
	SELECT b.id, b.user_id, p.display_name, b.name, b.description
	FROM benches b
	JOIN profiles p
		ON p.user_id = b.user_id
	WHERE b.user_id = ?
		AND b.deleted_at IS NULL
	ORDER BY b.updated_at;`

// Explore modules page
func Explore(c *gin.Context) {
	var p []ExploreBench

	result := model.DB.Raw(exploreQuery).Scan(&p)
	if result.Error != nil {
		zap.L().Panic("could not run explore query", zap.Error(result.Error))
	}

	m := map[string]interface{}{
		"Benches": p,
	}

	view.RenderTemplate(c, "bench/explore.tmpl", "EDeA - Explore Benches", m)
}

// ExploreOwn modules page
func ExploreOwn(c *gin.Context) {
	var p []ExploreBench
	user := c.Keys["user"].(*model.User)

	result := model.DB.Raw(exploreOwnQuery, user.ID).Scan(&p)
	if result.Error != nil {
		zap.L().Panic("could not run explore query", zap.Error(result.Error))
	}

	m := map[string]interface{}{
		"Benches": p,
	}

	view.RenderTemplate(c, "bench/explore_own.tmpl", "EDeA - Explore your Benches", m)
}
