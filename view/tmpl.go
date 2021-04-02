package view

// SPDX-License-Identifier: EUPL-1.2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
)

var (
	tmplFunctions = map[string]interface{}{
		"icon": Icon,
	}
)

// Icon returns the html code (as inline svg) for a icon from our icon library
// this function gets called by the templates
func Icon(name string) (html string, err error) {
	b, err := backend.BootstrapIcons.ReadFile(fmt.Sprintf("static/icons/%s.svg", name))
	if err != nil {
		return "", err
	}
	html = string(b)
	return
}

// RenderTemplate renders a go template
func RenderTemplate(ctx context.Context, fn string, data map[string]interface{}, w io.Writer) {
	tmplFile := filepath.Join(tmplPath, fn)

	if data == nil {
		data = make(map[string]interface{})
	}

	data["Dev"] = config.Cfg.Dev

	u, ok := ctx.Value(util.UserContextKey).(*model.User)
	if ok {
		data["User"] = u
		var moduleCount int64
		result := model.DB.
			Model(&model.Bench{}).
			Joins(`JOIN bench_modules bm ON "benches".id = bm.bench_id`).
			Where(`"benches".user_id = ? AND "benches".active = true AND bm.deleted_at IS NULL`, u.ID).
			Count(&moduleCount)

		if result.Error != nil {
			log.Panic().Err(result.Error).Msg("could not query active bench module count")
		} else {
			data["BenchModCount"] = moduleCount
		}
	}

	// create an empty template to associate our own functions to
	t := template.New("base").Funcs(tmplFunctions)

	// parse template fragments
	template.Must(t.ParseGlob(fmt.Sprintf("%s/fragments/*.tmpl", tmplPath)))

	// parse the template to render separately and add it with the name given
	// this is a workaround so that we don't end up with naming conflicts with
	// templates in different folders.
	// WARN: the fragments need to retain unique names though as we strip the folder
	// 		 prefix of them. this would only lead to issues though if there is a naming
	//		 conflict between the top folder and the fragments (both a page and fragment
	//       named index.tmpl).
	b, err := os.ReadFile(tmplFile)
	if err != nil {
		log.Panic().Err(err).Msgf("could not read template: %s", tmplFile)
	}
	tmpl := t.New(fn)
	if _, err := tmpl.Parse(string(b)); err != nil {
		log.Panic().Err(err).Msgf("could not parse template %s", fn)
	}

	// run our template with the data to render and the fragments
	if err := t.ExecuteTemplate(w, fn, data); err != nil {
		log.Panic().Err(err).Msgf("failed to render template: %s", err)
	}
}

// Template returns a http.HandlerFunc to render a specific template w/o further parameters
func Template(tmplName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		RenderTemplate(r.Context(), tmplName, nil, w)
	}
}

// RenderErrTemplate renders a page with error information
func RenderErrTemplate(ctx context.Context, w http.ResponseWriter, tmpl string, err error) {
	data := map[string]interface{}{
		"Error": err.Error(),
	}

	RenderTemplate(ctx, tmpl, data, w)
}
