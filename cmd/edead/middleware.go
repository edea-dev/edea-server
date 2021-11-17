package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/auth"
)

func middleware(r *gin.Engine) {
	r.Use(
		auth.Authenticate(),
	)
}
