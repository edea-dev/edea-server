package view

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edea-server/internal/model"
)

// CurrentUser returns the full User object when logged in or nil otherwise
func CurrentUser(c *gin.Context) *model.User {
	u, ok := c.Keys["user"].(*model.User)
	if !ok {
		return nil
	}

	return u
}
