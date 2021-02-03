package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"gitlab.com/edea-dev/edea/backend/auth"
	"gitlab.com/edea-dev/edea/backend/view"
	"gitlab.com/edea-dev/edea/backend/view/bench"
	"gitlab.com/edea-dev/edea/backend/view/module"
	"gitlab.com/edea-dev/edea/backend/view/user"
)

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/img/favicon.ico")
}

func routes(r *mux.Router, provider auth.Provider) {
	r.HandleFunc("/", view.Template("index.tmpl"))                                             // index
	r.HandleFunc("/about", view.Template("about.tmpl"))                                        // about EDeA
	r.HandleFunc("/explore", module.Explore)                                                   // explore modules
	r.Handle("/module/new", auth.Middleware(http.HandlerFunc(module.New))).Methods("GET")      // new module page
	r.Handle("/module/new", auth.Middleware(http.HandlerFunc(module.Create))).Methods("POST")  // add new module
	r.Handle("/module/{id}", auth.Middleware(http.HandlerFunc(module.Update))).Methods("POST") // view new module or adjust params
	r.Handle("/module/{id}", auth.Middleware(http.HandlerFunc(module.View))).Methods("GET")    // view module

	r.Handle("/bench/current", auth.Middleware(http.HandlerFunc(bench.Current))).Methods("GET")    // view current bench
	r.Handle("/bench/new", auth.Middleware(http.HandlerFunc(bench.New))).Methods("GET")            // new bench form
	r.Handle("/bench/new", auth.Middleware(http.HandlerFunc(bench.Create))).Methods("POST")        // add a new bench
	r.Handle("/bench/{id}", auth.Middleware(http.HandlerFunc(bench.Update))).Methods("POST")       // update a bench
	r.Handle("/bench/{id}", auth.Middleware(http.HandlerFunc(bench.View))).Methods("GET")          // view a bench
	r.Handle("/bench/add/{id}", auth.Middleware(http.HandlerFunc(bench.AddModule))).Methods("GET") // add a module to the active bench
	r.Handle("/bench/remove/{id}", auth.Middleware(http.HandlerFunc(bench.RemoveModule))).Methods("GET")

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
	r.Handle("/profile", auth.Middleware(http.HandlerFunc(user.Profile))).Methods("GET")
	r.Handle("/profile", auth.Middleware(http.HandlerFunc(user.UpdateProfile))).Methods("POST")

	r.Handle("/callback", auth.Middleware(http.HandlerFunc(provider.CallbackHandler)))
	r.HandleFunc("/login", provider.LoginHandler)
	r.HandleFunc("/logout", provider.LogoutHandler)
}
