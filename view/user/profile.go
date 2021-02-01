package user

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
)

// Profile displays the user data
func Profile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(model.AuthContextKey).(model.AuthClaims)
	u := model.User{AuthUUID: claims.Subject}

	if result := model.DB.First(&u); result.Error != nil {
		log.Error().Err(result.Error).Msgf("could not fetch user data for %s", claims.Subject)
	}

	p := model.Profile{UserID: u.ID}

	if result := model.DB.Where(&p).First(&p); result.Error != nil {
		log.Error().Err(result.Error).Msgf("could not fetch user data for %s", claims.Subject)
	}

	// TODO: fetch profile data from cache, or more data to display
	data := map[string]interface{}{
		"User":    u,
		"Profile": p,
	}

	view.RenderMarkdown("user/profile.md", data, w)
}

// UpdateProfile updates the user data
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "user/profile.md", err)
		return
	}
	// update the id of the current user only
	u := view.CurrentUser(r)

	profile := new(model.Profile)
	if err := util.FormDecoder.Decode(profile, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "user/profile.md", err)
		return
	}

	profile.UserID = u.ID

	result := model.DB.WithContext(r.Context()).Save(profile)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not update profile")
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
