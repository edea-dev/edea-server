package main

// SPDX-License-Identifier: EUPL-1.2

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"gitlab.com/edea-dev/edea/backend/auth"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/search"
	"gitlab.com/edea-dev/edea/backend/view"
	"gitlab.com/edea-dev/edea/backend/view/bench"
	"gitlab.com/edea-dev/edea/backend/view/module"
	"gitlab.com/edea-dev/edea/backend/view/user"
)

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/img/favicon.ico")
}

func routes(r *mux.Router) {
	r.HandleFunc("/", view.Template("index.tmpl", "EDeA"))              // index
	r.HandleFunc("/about", view.Template("about.tmpl", "EDeA - About")) // about EDeA

	r.HandleFunc("/search", view.Template("search.tmpl", "EDeA - Search")) // Search page

	r.Handle("/module/new", auth.RequireAuth(http.HandlerFunc(module.New))).Methods("GET")            // new module page
	r.Handle("/module/new", auth.RequireAuth(http.HandlerFunc(module.Create))).Methods("POST")        // add new module
	r.HandleFunc("/module/explore", module.Explore).Methods("GET")                                    // explore public modules
	r.HandleFunc("/module/user/{id}", module.ExploreUser).Methods("GET")                              // view a users modules
	r.Handle("/module/{id}", auth.RequireAuth(http.HandlerFunc(module.Update))).Methods("POST")       // view new module or adjust params
	r.HandleFunc("/module/{id}", module.View).Methods("GET")                                          // view module
	r.Handle("/module/delete/{id}", auth.RequireAuth(http.HandlerFunc(module.Delete))).Methods("GET") // delete module
	r.Handle("/module/pull/{id}", auth.RequireAuth(http.HandlerFunc(module.Pull))).Methods("GET")     // pull repo of module
	r.HandleFunc("/module/history/{id}", module.ViewHistory).Methods("GET")                           // show revision history of a module
	r.HandleFunc("/module/diff/{id}", module.Diff).Methods("GET")

	r.Handle("/bench/current", auth.RequireAuth(http.HandlerFunc(bench.Current))).Methods("GET")                 // view current bench
	r.Handle("/bench/new", auth.RequireAuth(view.Template("bench/new.tmpl", "EDeA - New Bench"))).Methods("GET") // new bench form
	r.Handle("/bench/new", auth.RequireAuth(http.HandlerFunc(bench.Create))).Methods("POST")                     // add a new bench
	r.HandleFunc("/bench/explore", bench.Explore).Methods("GET")                                                 // explore public workbenches
	r.Handle("/bench/{id}", auth.RequireAuth(http.HandlerFunc(bench.Update))).Methods("POST")                    // update a bench
	r.HandleFunc("/bench/{id}", bench.View).Methods("GET")                                                       // view a bench
	r.Handle("/bench/update/{id}", auth.RequireAuth(http.HandlerFunc(bench.ViewUpdate))).Methods("GET")          // update form view of a bench
	r.Handle("/bench/add/{id}", auth.RequireAuth(http.HandlerFunc(bench.AddModule))).Methods("GET")              // add a module to the active bench
	r.Handle("/bench/remove/{id}", auth.RequireAuth(http.HandlerFunc(bench.RemoveModule))).Methods("GET")        // remove module from workbench
	r.Handle("/bench/delete/{id}", auth.RequireAuth(http.HandlerFunc(bench.Delete))).Methods("GET")              // delete the workbench
	r.HandleFunc("/bench/user/{id}", bench.ListUser).Methods("GET")                                              // list workbenches of a specific user
	r.Handle("/bench/fork/{id}", auth.RequireAuth(http.HandlerFunc(bench.Fork))).Methods("GET")                  // fork a workbench
	r.Handle("/bench/activate/{id}", auth.RequireAuth(http.HandlerFunc(bench.SetActive))).Methods("GET")         // set a workbench as active
	r.HandleFunc("/bench/merge/{id}", bench.Merge).Methods("GET")

	r.HandleFunc("/favicon.ico", faviconHandler)
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/profile", pprof.Trace)

	// api routes
	//r.HandleFunc("/api/module", api.REST(&api.Module{}))
	//r.HandleFunc("/api/user", api.REST(&api.User{}))
	//r.HandleFunc("/api/bench", api.REST(&api.Bench{}))

	// static files
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./static/css/"))))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./static/js/"))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("./static/img/"))))
	r.PathPrefix("/fonts/").Handler(http.StripPrefix("/fonts/", http.FileServer(http.Dir("./static/fonts/"))))
	r.PathPrefix("/icons/").Handler(http.StripPrefix("/icons/", http.FileServer(http.Dir("./static/icons/"))))

	// TODO: let our IAP do that
	r.Handle("/profile", auth.RequireAuth(http.HandlerFunc(user.Profile))).Methods("GET")
	r.Handle("/profile", auth.RequireAuth(http.HandlerFunc(user.UpdateProfile))).Methods("POST")
	r.Handle("/profile/export", auth.RequireAuth(http.HandlerFunc(user.DataExport))).Methods("GET")

	r.HandleFunc("/callback", auth.CallbackHandler)
	r.HandleFunc("/logout_callback", auth.LogoutCallbackHandler)
	r.HandleFunc("/login", auth.LoginHandler)
	r.HandleFunc("/logout", auth.LogoutHandler)

	r.HandleFunc("/search/_bulk_update", search.ReIndexDB)

	// the login action redirects to the OIDC provider, with mock auth we have to provide this ourselves
	if config.Cfg.Auth.UseMock {
		r.HandleFunc("/auth", auth.LoginFormHandler).Methods("GET")
		r.HandleFunc("/auth", auth.LoginPostHandler).Methods("POST")
		r.HandleFunc("/.well-known/openid-configuration", auth.WellKnown)
		r.HandleFunc("/keys", auth.Keys)
		r.HandleFunc("/userinfo", auth.Userinfo)
		r.HandleFunc("/token", auth.Token).Methods("POST")
	}
}
