package user

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"

	"gitlab.com/edea-dev/edead/internal/model"
	"gitlab.com/edea-dev/edead/internal/util"
	"gitlab.com/edea-dev/edead/internal/view"
	"go.uber.org/zap"
)

// Profile displays the user data
func Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := ctx.Value(util.UserContextKey).(*model.User)

	p := model.Profile{UserID: u.ID}

	if result := model.DB.Where(&p).First(&p); result.Error != nil {
		zap.L().Panic("could not fetch profile data", zap.Error(result.Error), zap.String("subject", u.AuthUUID))
	}

	// TODO: fetch profile data from cache, or more data to display
	data := map[string]interface{}{
		"Profile": p,
	}

	view.RenderTemplate(ctx, "profile.tmpl", "EDeA - Profile", data, w)
}

// UpdateProfile updates the user data
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErrTemplate(r.Context(), w, "user/profile.tmpl", err)
		return
	}
	// update the id of the current user only
	ctx := r.Context()
	u := ctx.Value(util.UserContextKey).(*model.User)

	profile := new(model.Profile)
	if err := util.FormDecoder.Decode(profile, r.Form); err != nil {
		view.RenderErrTemplate(ctx, w, "user/profile.tmpl", err)
		return
	}

	profile.UserID = u.ID

	result := model.DB.WithContext(ctx).Save(profile)
	if result.Error != nil {
		zap.L().Panic("could not update profile", zap.Error(result.Error))
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
