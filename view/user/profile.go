package user

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
)

// Profile displays the user data
func Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := ctx.Value(util.UserContextKey).(*model.User)

	p := model.Profile{UserID: u.ID}

	if result := model.DB.Where(&p).First(&p); result.Error != nil {
		log.Panic().Err(result.Error).Msgf("could not fetch profile data for sub %s", u.AuthUUID)
	}

	// TODO: fetch profile data from cache, or more data to display
	data := map[string]interface{}{
		"Profile": p,
	}

	view.RenderTemplate(ctx, "profile.tmpl", data, w)
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
		log.Panic().Err(result.Error).Msg("could not update profile")
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
