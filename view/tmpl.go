package view

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
)

// RenderTemplate renders a go template
func RenderTemplate(fn string, data map[string]interface{}, w io.Writer) {
	tmplFile := fmt.Sprintf("%s%s", tmplPath, fn)

	data["Dev"] = true

	// parse base templates
	t := template.Must(template.ParseGlob(fmt.Sprintf("%s/fragments/*.tmpl", tmplPath)))

	// read the template to render
	t = template.Must(t.ParseFiles(tmplFile))

	if err := t.ExecuteTemplate(w, fn, data); err != nil {
		log.Panic().Err(err).Msgf("failed to render template: %s", err)
	}
}

// Template returns a http handler for templates which only need the current user information
func Template(tmplName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// vars := mux.Vars(r)
		user := CurrentUser(r)

		// all packed up,
		m := map[string]interface{}{
			"User":  user,
			"Error": nil,
		}

		// and ready to go
		RenderTemplate(tmplName, m, w)
	}
}

// RenderErrTemplate renders a page with error information
func RenderErrTemplate(ctx context.Context, w http.ResponseWriter, tmpl string, err error) {
	user := ctx.Value(util.UserContextKey).(*model.User)
	data := map[string]interface{}{
		"User":  user,
		"Error": err.Error(),
	}

	RenderTemplate(tmpl, data, w)
}
