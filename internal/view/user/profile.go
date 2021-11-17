package user

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/model"
	"gitlab.com/edea-dev/edead/internal/util"
	"gitlab.com/edea-dev/edead/internal/view"
	"go.uber.org/zap"
)

// Profile displays the user data
func Profile(c *gin.Context) {
	u := c.Value(util.UserContextKey).(*model.User)

	p := model.Profile{UserID: u.ID}

	if result := model.DB.Where(&p).First(&p); result.Error != nil {
		zap.L().Panic("could not fetch profile data", zap.Error(result.Error), zap.String("subject", u.AuthUUID))
	}

	// TODO: fetch profile data from cache, or more data to display
	data := map[string]interface{}{
		"Profile": p,
	}

	view.RenderTemplate(c, "profile.tmpl", "EDeA - Profile", data)
}

// UpdateProfile updates the user data
func UpdateProfile(c *gin.Context) {
	// update the id of the current user only
	u := c.Value(util.UserContextKey).(*model.User)

	profile := new(model.Profile)
	if err := c.Bind(profile); err != nil {
		view.RenderErrTemplate(c, "user/profile.tmpl", err)
		return
	}

	profile.UserID = u.ID

	result := model.DB.WithContext(c).Save(profile)
	if result.Error != nil {
		zap.L().Panic("could not update profile", zap.Error(result.Error))
	}

	c.Redirect(http.StatusSeeOther, "/profile")
}
