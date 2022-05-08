package view

// SPDX-License-Identifier: EUPL-1.2

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/gin-gonic/gin"
	edea "gitlab.com/edea-dev/edea-server"
	"gitlab.com/edea-dev/edea-server/internal/config"
	"gitlab.com/edea-dev/edea-server/internal/model"
	"go.uber.org/zap"
)

var (
	tmplFunctions = map[string]interface{}{
		"icon": Icon,
	}
)

// Icon returns the html code (as inline svg) for a icon from our icon library
// this function gets called by the templates
func Icon(name string) (html string, err error) {
	b, err := edea.BootstrapIcons.ReadFile(fmt.Sprintf("static/icons/%s.svg", name))
	if err != nil {
		return "", err
	}
	html = string(b)
	return
}

// RenderTemplate renders a go template
func RenderTemplate(c *gin.Context, fn, title string, data map[string]interface{}) {
	tmplFile := filepath.Join(tmplPath, fn)

	if data == nil {
		data = make(map[string]interface{})
	}

	data["Dev"] = config.Cfg.Dev
	if _, ok := data["Title"]; !ok {
		data["Title"] = title
	}

	u, ok := c.Keys["user"].(*model.User)
	if ok {
		data["User"] = u
		var moduleCount int64
		result := model.DB.
			Model(&model.Bench{}).
			Joins(`JOIN bench_modules bm ON "benches".id = bm.bench_id`).
			Where(`"benches".user_id = ? AND "benches".active = true AND bm.deleted_at IS NULL`, u.ID).
			Count(&moduleCount)

		if result.Error != nil {
			zap.L().Panic("could not query active bench module count", zap.Error(result.Error))
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
		zap.L().Panic("could not read template", zap.Error(err), zap.String("file", tmplFile))
	}
	tmpl := t.New(fn)
	if _, err := tmpl.Parse(string(b)); err != nil {
		zap.L().Panic("could not parse template", zap.Error(err), zap.String("file_name", fn))
	}

	// run our template with the data to render and the fragments
	if err := t.ExecuteTemplate(c.Writer, fn, data); err != nil {
		zap.L().Panic("failed to render template", zap.Error(err))
	}
}

// Template returns a http.HandlerFunc to render a specific template w/o further parameters
func Template(tmplName, title string) gin.HandlerFunc {
	return func(r *gin.Context) {
		RenderTemplate(r, tmplName, title, nil)
	}
}

// TemplateM returns a http.HandlerFunc to render a specific template w/ a statically allocated map as parameters
func TemplateM(tmplName string, m map[string]interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		RenderTemplate(c, tmplName, "", m)
	}
}

// RenderErrTemplate renders a page with error information
func RenderErrTemplate(c *gin.Context, tmpl string, err error) {
	data := map[string]interface{}{
		"Error": err.Error(),
	}

	RenderTemplate(c, tmpl, "EDeA - Error", data)
}
