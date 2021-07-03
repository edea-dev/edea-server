package user

// SPDX-License-Identifier: EUPL-1.2

import (
	"archive/zip"
	"bytes"
	"log"
	"net/http"

	"gitlab.com/edea-dev/edead/model"
	"gitlab.com/edea-dev/edead/util"
	"gitlab.com/edea-dev/edead/view"
	"gopkg.in/yaml.v3"
)

// DataExport provides the user a zip file with their personal data
//     This should contain any GDPR relevant data as well as their projects,
//     modules, benches, etc.
// TODO: test it
func DataExport(w http.ResponseWriter, r *http.Request) {
	var benches []model.Bench
	var modules []model.Module
	var profile model.Profile

	ctx := r.Context()
	u := ctx.Value(util.UserContextKey).(*model.User)

	// load a users benches
	result := model.DB.Model(&model.Bench{}).Preload("Modules").Where("user_id = ?", u.ID).Find(benches)
	if result.Error != nil {
		view.RenderErrMarkdown(ctx, w, "user/404.md", result.Error)
		return
	}

	// load modules
	result = model.DB.Model(&model.Module{}).Preload("Category").Where("user_id = ?", u.ID).Find(modules)
	if result.Error != nil {
		view.RenderErrMarkdown(ctx, w, "user/404.md", result.Error)
		return
	}

	// profile info
	result = model.DB.Model(&model.Profile{}).Where("user_id = ?", u.ID).Find(profile)
	if result.Error != nil {
		view.RenderErrMarkdown(ctx, w, "user/404.md", result.Error)
		return
	}

	// encode all the data to yaml
	b1, err1 := yaml.Marshal(benches)
	b2, err2 := yaml.Marshal(modules)
	b3, err3 := yaml.Marshal(profile)

	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal(err1, err2, err3)
	}

	// build an archive with the data
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	var files = []struct {
		Name string
		Body []byte
	}{
		{"benches.yml", b1},
		{"modules.yml", b2},
		{"profile.yml", b3},
	}
	for _, file := range files {
		f, err := zw.Create(file.Name)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			log.Fatal(err)
		}
	}

	err := zw.Close()
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/zip")
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Fatal(err)
	}
}
