package bench

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/view"
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

// Explore modules page
func Explore(w http.ResponseWriter, r *http.Request) {
	var p []ExploreBench

	result := model.DB.Raw(exploreQuery).Scan(&p)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not run explore query")
	}

	m := map[string]interface{}{
		"Benches": p,
	}

	view.RenderTemplate(r.Context(), "bench/explore.tmpl", m, w)
}
