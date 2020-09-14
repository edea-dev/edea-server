package user

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/auth"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/util"
	"gitlab.com/edea-dev/edea/backend/view"
)

// Profile displays the user data
func Profile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKey).(auth.Claims)
	u := model.User{AuthUUID: claims.Subject}

	if err := model.DB.Model(&u).Column("user.*").Relation("Profile").Select(); err != nil {
		log.Error().Err(err).Msgf("could not fetch user data for %s", claims.Subject)
	}

	// TODO: fetch profile data from cache, or more data to display
	data := map[string]interface{}{
		"User": u,
	}

	view.RenderMarkdown("user/profile.md", data, w)
}

// UpdateProfile updates the user data
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		view.RenderErr(r.Context(), w, "user/profile.md", err)
		return
	}

	profile := new(model.Profile)
	if err := util.FormDecoder.Decode(profile, r.Form); err != nil {
		view.RenderErr(r.Context(), w, "user/profile.md", err)
		return
	}

	user := r.Context().Value(util.UserContextKey).(model.User)

	if profile.UserID != user.ID {
		if user.IsAdmin {
			// TODO: admin changing stuff
		} else {
			view.RenderErr(r.Context(), w, "user/profile.md", util.ErrImSorryDave)
			return
		}
	}

	_, err := model.DB.Model(profile).Update()
	if err != nil {
		log.Panic().Err(err).Msg("could not update profile")
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
