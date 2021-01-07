package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

var addr = flag.String("addr", "localhost:8282", "http service address")
var logger = log.New("Server")
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	configPath := flag.String("config-path", os.Getenv("EBIDLOCAL_CONFIG"), "path to the config file.")
	logger.Info.Println(configPath)
	flag.Parse()

	registerHTTPHandlers()

	logger.Info.Println("Listening on http://" + *addr)
	logger.Error.Fatal(http.ListenAndServe(*addr, nil))
}

func registerHTTPHandlers() {
	r := mux.NewRouter()

	registerUserRoutes(r.PathPrefix("/user").Subrouter())

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static")))
	http.Handle("/", r)
}

func registerUserRoutes(router *mux.Router) *mux.Router {
	//router.Methods("POST").HandlerFunc(createUser)
	// router.HandleFunc("/{userID}/", func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	logger.Info.Println("In user", vars["userID"])
	// })

	// router.Path("/{userID}/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	logger.Info.Println("In user", vars["userID"])
	// })
	router.Path("/{userID}/").Handler(http.StripPrefix("/user/", http.FileServer(http.Dir("./web/user")))).
		Name("userDir")

	return router
}

func createUser(w http.ResponseWriter, r *http.Request) {
	logger.Info.Println("Should handle POST")
}

//http.Handle("/", http.FileServer(http.Dir("./web/static")))
// r.HandleFunc("/", HomeHandler)
// r.HandleFunc("/products", ProductsHandler)
// r.HandleFunc("/articles", ArticlesHandler)
// r.HandleFunc("/products/{key}", ProductHandler)
// r.HandleFunc("/articles/{category}/", ArticlesCategoryHandler)
// r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler)

// http.HandleFunc("/forcegraph", func(w http.ResponseWriter, r *http.Request) {
// 	session, _ := store.Get(r, "uid")
// 	if _, ok := session.Values["username"]; !ok {
// 		session.Values["username"] = "No Name"
// 	}

// 	err := session.Save(r, w)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		logger.Error.Println(err)
// 		return
// 	}

// 	graph.RegisterConnection(conn, session.ID, session.Values["username"].(string))
// })
