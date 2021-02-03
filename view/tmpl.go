package view

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
)

// RenderTemplate renders a go template
func RenderTemplate(fn string, data map[string]interface{}, w io.Writer) {
	tmplFile := filepath.Join(tmplPath, fn)

	data["Dev"] = true

	// parse base templates
	t := template.Must(template.ParseGlob(fmt.Sprintf("%s/fragments/*.tmpl", tmplPath)))

	// parse the template to render separately and add it with the name given
	// this is a workaround so that we don't end up with naming conflicts with
	// templates in different folders.
	// WARN: the fragments need to retain unique names though as we strip the folder
	// 		 prefix of them. this would only lead to issues though if there is a naming
	//		 conflict between the top folder and the fragments (both a page and fragment
	//       named index.tmpl).
	tp := template.Must(template.ParseFiles(tmplFile))
	t.AddParseTree(fn, tp.Tree)

	// run our template with the data to render and the fragments
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
