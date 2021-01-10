package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func New(config Config, ebidlocal *ebidlocal.Ebidlocal) *Server {
	var server = Server{
		config:    config,
		logger:    log.New("Server"),
		ebidlocal: ebidlocal,
	}

	if url, err := url.Parse(fmt.Sprintf("%s:%d", config.Address, config.Port)); err != nil {
		server.logger.Error.Fatalln(err)
	} else {
		server.addr = url
	}

	server.registerHTTPHandlers()

	return &server
}

type Server struct {
	logger    *log.Logger
	addr      *url.URL
	ebidlocal *ebidlocal.Ebidlocal
	config    Config
}

func (s *Server) Run() {
	s.logger.Info.Printf("Listening on %s\n", s.addr.String())
	s.logger.Error.Fatal(http.ListenAndServe(s.addr.String(), nil))
}

func (s *Server) registerHTTPHandlers() {
	r := mux.NewRouter()

	s.registerUserRoutes(r.PathPrefix("/user").Subrouter())

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static")))

	loggedRouter := handlers.RecoveryHandler()(handlers.LoggingHandler(os.Stdout, r))
	http.Handle("/", loggedRouter)
}

func (s *Server) registerUserRoutes(router *mux.Router) *mux.Router {
	router.Methods("POST").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.createUser), "application/json"))
	router.Path("/{userID}/").Handler(http.StripPrefix("/user/", http.FileServer(http.Dir("./web/user")))).Name("userDir")

	return router

	// router.HandleFunc("/{userID}/", func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	logger.Info.Println("In user", vars["userID"])
	// })

	// router.Path("/{userID}/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	logger.Info.Println("In user", vars["userID"])
	// })
}

//User user data.
type User struct {
	UserName string `json:"username"`
}

func (u User) String() string {
	return fmt.Sprintf("User name: '%s'", u.UserName)
}

//IsValid validate user data.
func (u User) IsValid() bool {
	return u.UserName != ""
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !user.IsValid() {
		respondError(w, http.StatusBadRequest, "User name is required.")
		return
	}

	s.logger.Info.Println(user)
	respondJSON(w, http.StatusCreated, user)
}

// respondJSON makes the response with payload as json format
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// respondError makes the error response with payload as json format
func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
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
