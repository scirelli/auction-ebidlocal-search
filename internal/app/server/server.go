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
	router.Path("/{userID}/watchlist").Methods("POST").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.createWatchlist), "application/json")).Name("createWatchlist")
	router.Methods("POST").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.createUser), "application/json"))

	router.PathPrefix("/{userID}/data.json").Handler(http.StripPrefix("/user/", http.FileServer(http.Dir("./web/user")))).Name("userData")

	router.PathPrefix("/{userID}/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := mux.Vars(r)["userID"]
		d := fmt.Sprintf("./web/user/%s/static", userID)
		rm := fmt.Sprintf("/user/%s/", userID)
		http.StripPrefix(rm, http.FileServer(http.Dir(d))).ServeHTTP(w, r)
	})).Name("userDir")

	return router
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User
	var err error
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !user.IsValid() {
		respondError(w, http.StatusBadRequest, "User name is required.")
		return
	}

	user.ID, err = s.ebidlocal.CreateUser(user.Name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Could not create user")
		return
	}
	s.logger.Info.Println(user.ID)

	w.Header().Set("Location", fmt.Sprintf("/user/%s/", url.PathEscape(user.ID)))
	respondJSON(w, http.StatusCreated, user)
}

func (s *Server) createWatchlist(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	s.logger.Info.Printf("Create watch list called '%s'", userID)
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
