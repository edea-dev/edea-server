package auth

// SPDX-License-Identifier: EUPL-1.2

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/model"
	"go.uber.org/zap"
)

// Provider interface to be implemented by Identity Providers
type Provider interface {
	CallbackHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
	LogoutCallbackHandler(w http.ResponseWriter, r *http.Request)
	Init() error
}

var (
	verifier *oidc.IDTokenVerifier
)

func processAuth(c *gin.Context) error {
	auth := c.GetHeader("Authorization")
	raw, err := c.Cookie("jwt")

	if err != nil && len(auth) == 0 {
		return model.ErrUnauthorized
	}

	if len(auth) > 0 {
		raw = strings.Replace(auth, "Bearer ", "", 1)
	}

	claims := model.AuthClaims{}

	// verify claims
	idToken, err := verifier.Verify(c, raw)
	if err != nil {
		zap.L().Error("could not verify jwt", zap.Error(err))

		// remove offending jwt cookie
		cookie := http.Cookie{
			Name:     "jwt",
			Value:    "",
			Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(c.Writer, &cookie)

		return err
	}

	if err := idToken.Claims(&claims); err != nil {
		// claims aren't there?
		return err
	}

	// get the current user object from the database
	user := &model.User{AuthUUID: claims.Subject}
	result := model.DB.Model(user).Where(user).First(user)
	if result.Error != nil {
		return fmt.Errorf("could not fetch user data for %s (%v)", claims.Subject, result.Error)
	}

	// add claims and user object to the context
	c.Keys = make(map[string]interface{})
	c.Keys["auth"] = claims
	c.Keys["user"] = user

	return nil
}

// RequireAuth checks if there is a valid json web token in the request
func RequireAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if _, ok := c.Keys["auth"]; !ok {
			c.AbortWithError(
				http.StatusUnauthorized,
				errors.New("Authorization header/session cookie missing"),
			)
			return
		}

		// auth key is set, everything is fine
		c.Next()
	})
}

// Authenticate checks if an authorization header or cookie is present and processes it
func Authenticate() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if err := processAuth(c); err != nil {
			// only show an error if something is wrong with the token (expired tokens are not an error)
			if !errors.Is(err, model.ErrUnauthorized) && !strings.Contains(err.Error(), "expired") {
				zap.L().Error("could not process authentication cookie/header", zap.Error(err))
				c.AbortWithError(http.StatusInternalServerError, err)
			}
		}
		c.Next()
	})
}

func createUser(claims *model.AuthClaims) {
	u := model.User{
		AuthUUID: claims.Subject,
		Handle:   claims.Subject,
	}

	// set
	if claims.Nickname != "" {
		u.Handle = claims.Nickname
	}

	if result := model.DB.Model(&u).Create(&u); result.Error != nil {
		zap.L().Error("could not create new user", zap.Error(result.Error), zap.String("auth_uuid", claims.Subject))
	}

	p := model.Profile{DisplayName: claims.Nickname, Avatar: claims.Picture, UserID: u.ID}

	if result := model.DB.Model(&p).Create(&p); result.Error != nil {
		zap.L().Panic("could not create new user", zap.Error(result.Error), zap.String("auth_uuid", claims.Subject))
	}

	zap.L().Info("created a new user", zap.Object("user", &u))
}
