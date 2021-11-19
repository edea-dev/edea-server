package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edead/internal/auth"
	"gitlab.com/edea-dev/edead/internal/config"
	"gitlab.com/edea-dev/edead/internal/search"
	"gitlab.com/edea-dev/edead/internal/view"
	"gitlab.com/edea-dev/edead/internal/view/bench"
	"gitlab.com/edea-dev/edead/internal/view/module"
	"gitlab.com/edea-dev/edead/internal/view/user"
)

func faviconHandler(c *gin.Context) {
	c.File("./static/img/favicon.ico")
}

func routes(router *gin.Engine) {
	r := router.Group("/", auth.Authenticate())
	r.GET("/", view.Template("index.tmpl", "EDeA"))              // index
	r.GET("/about", view.Template("about.tmpl", "EDeA - About")) // about EDeA

	r.GET("/search", view.Template("search.tmpl", "EDeA - Search")) // Search page

	a := r.Group("/", auth.RequireAuth())

	a.GET("/module/new", module.New)                  // new module page
	a.POST("/module/new", module.Create)              // add new module
	r.GET("/module/explore", module.Explore)          // explore public modules
	r.GET("/module/user/{id}", module.ExploreUser)    // view a users modules
	a.POST("/module/{id}", module.Update)             // view new module or adjust params
	r.GET("/module/{id}", module.View)                // view module
	a.GET("/module/delete/{id}", module.Delete)       // delete module
	a.GET("/module/pull/{id}", module.Pull)           // pull repo of module
	r.GET("/module/history/{id}", module.ViewHistory) // show revision history of a module
	r.GET("/module/diff/{id}", module.Diff)
	a.GET("/module/build_book/{id}", module.BuildBook)

	a.GET("/bench/current", bench.Current)                                   // view current bench
	a.GET("/bench/new", view.Template("bench/new.tmpl", "EDeA - New Bench")) // new bench form
	a.POST("/bench/new", bench.Create)                                       // add a new bench
	r.GET("/bench/explore", bench.Explore)                                   // explore public workbenches
	a.POST("/bench/{id}", bench.Update)                                      // update a bench
	r.GET("/bench/{id}", bench.View)                                         // view a bench
	a.GET("/bench/update/{id}", bench.ViewUpdate)                            // update form view of a bench
	a.GET("/bench/add/{id}", bench.AddModule)                                // add a module to the active bench
	a.GET("/bench/remove/{id}", bench.RemoveModule)                          // remove module from workbench
	a.GET("/bench/delete/{id}", bench.Delete)                                // delete the workbench
	r.GET("/bench/user/{id}", bench.ListUser)                                // list workbenches of a specific user
	a.GET("/bench/fork/{id}", bench.Fork)                                    // fork a workbench
	a.GET("/bench/activate/{id}", bench.SetActive)                           // set a workbench as active
	r.GET("/bench/merge/{id}", bench.Merge)

	r.GET("/favicon.ico", faviconHandler)

	// api routes
	//r.HandleFunc("/api/module", api.REST(&api.Module{}))
	//r.HandleFunc("/api/user", api.REST(&api.User{}))
	//r.HandleFunc("/api/bench", api.REST(&api.Bench{}))

	// static files
	router.Static("/css", "./static/css")
	router.Static("/js", "./static/js")
	router.Static("/img", "./static/img")
	router.Static("/fonts", "./static/fonts")
	router.Static("/icons", "./static/icons")

	// mdbooks are built and served from here
	router.Static("/module/doc", config.Cfg.Cache.Book.Base)

	// TODO: let our IAP do that
	a.GET("/profile", user.Profile)
	a.POST("/profile", user.UpdateProfile)
	a.GET("/profile/export", user.DataExport)

	r.GET("/callback", auth.CallbackHandler)
	r.POST("/callback", auth.CallbackHandler)
	r.GET("/logout_callback", auth.LogoutCallbackHandler)
	r.GET("/login", auth.LoginHandler)
	r.POST("/logout", auth.LogoutHandler)

	a.GET("/search/_bulk_update", search.ReIndexDB)
	a.GET("/_module/_bulk_update", module.PullAllRepos)

	// the login action redirects to the OIDC provider, with mock auth we have to provide this ourselves
	if config.Cfg.Auth.UseMock {
		router.GET("/auth", auth.LoginFormHandler)
		router.POST("/auth", auth.LoginPostHandler)
		router.GET("/.well-known/openid-configuration", auth.WellKnown)
		router.GET("/keys", auth.Keys)
		router.GET("/userinfo", auth.Userinfo)
		router.POST("/token", auth.Token)
	}
}
