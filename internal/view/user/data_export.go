package user

// SPDX-License-Identifier: EUPL-1.2

import (
	"archive/zip"
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edea-server/internal/model"
	"gitlab.com/edea-dev/edea-server/internal/view"
	"gopkg.in/yaml.v3"
)

// DataExport provides the user a zip file with their personal data
//     This should contain any GDPR relevant data as well as their projects,
//     modules, benches, etc.
// TODO: test it
func DataExport(c *gin.Context) {
	var benches []model.Bench
	var modules []model.Module
	var profile model.Profile

	u := c.Keys["user"].(*model.User)

	// load a users benches
	result := model.DB.Model(&model.Bench{}).Preload("Modules").Where("user_id = ?", u.ID).Find(benches)
	if result.Error != nil {
		view.RenderErrMarkdown(c, "user/404.md", result.Error)
		return
	}

	// load modules
	result = model.DB.Model(&model.Module{}).Preload("Category").Where("user_id = ?", u.ID).Find(modules)
	if result.Error != nil {
		view.RenderErrMarkdown(c, "user/404.md", result.Error)
		return
	}

	// profile info
	result = model.DB.Model(&model.Profile{}).Where("user_id = ?", u.ID).Find(profile)
	if result.Error != nil {
		view.RenderErrMarkdown(c, "user/404.md", result.Error)
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

	c.Header("Content-Disposition", "attachment; filename=export.zip")
	http.ServeContent(c.Writer, c.Request, "export-zip", time.Now(), bytes.NewReader(buf.Bytes()))
}
