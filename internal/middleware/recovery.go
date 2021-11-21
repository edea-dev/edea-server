package middleware

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"
	"runtime/debug"
	"text/template"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(c *gin.Context, recovered interface{}) {
	c.Status(http.StatusInternalServerError)

	c.Request.ParseForm()
	m := map[string]interface{}{
		"keys":       c.Keys,
		"stacktrace": string(debug.Stack()),
		"error":      recovered,
		"route":      c.FullPath(),
		"vars":       c.Params,
		"form":       c.Request.Form,
		"headers":    c.Request.Header,
	}

	tmpl, err := template.New("error tmpl").Parse(devErrorTmpl)
	if err != nil {
		zap.L().Fatal(`error while parsing panic template 🤦‍♀️🤦🤦🤦🤦🤦`, zap.Error(err))
	}

	if err := tmpl.Execute(c.Writer, m); err != nil {
		zap.L().Fatal(`error while rendering panic template 🤦‍♀️🤦🤦🤦🤦🤦`, zap.Error(err))
	}

	// zap.L().Panic("recovery handler", zap.Error(err)) // TODO: add back r.Context()
}
