package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"gitlab.com/edea-dev/edea/backend/auth"
	"gitlab.com/edea-dev/edea/backend/view"
	"gitlab.com/edea-dev/edea/backend/view/project"
	"gitlab.com/edea-dev/edea/backend/view/user"
)

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/img/favicon.ico")
}

func routes(r *mux.Router, provider auth.Provider) {
	r.HandleFunc("/", view.Markdown("index.md"))                  // index
	r.HandleFunc("/about", view.Markdown("about.md"))             // about EDeA
	r.HandleFunc("/explore", project.Explore)                     // explore projects
	r.HandleFunc("/project/new", project.New).Methods("GET")      // new project page
	r.HandleFunc("/project/new", project.Create).Methods("POST")  // add new project
	r.HandleFunc("/project/{id}", project.Update).Methods("POST") // view new project or adjust params
	r.HandleFunc("/project/{id}", project.View).Methods("GET")    // view project

	r.HandleFunc("/favicon.ico", faviconHandler)
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/profile", pprof.Trace)

	// api routes
	//r.HandleFunc("/api/project", api.REST(&api.Project{}))
	//r.HandleFunc("/api/user", api.REST(&api.User{}))
	//r.HandleFunc("/api/bench", api.REST(&api.Bench{}))

	// static files
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./static/css/"))))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./static/js/"))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("./static/img/"))))

	// TODO: let our IAP do that
	r.Handle("/profile", auth.Middleware(http.HandlerFunc(user.Profile))).Methods("GET")
	r.Handle("/profile", auth.Middleware(http.HandlerFunc(user.UpdateProfile))).Methods("POST")

	r.Handle("/callback", auth.Middleware(http.HandlerFunc(provider.CallbackHandler)))
	r.HandleFunc("/login", provider.LoginHandler)
	r.HandleFunc("/logout", provider.LogoutHandler)
}
