package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal/watchlist"
	. "github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/store"
	storefs "github.com/scirelli/auction-ebidlocal-search/internal/app/server/store/fs"
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

	if config.ContentPath == "" {
		server.config.ContentPath = "."
	}
	if config.UserDir == "" {
		server.config.UserDir = filepath.Join(config.ContentPath, "web", "user")
		server.logger.Info.Printf("Defaulting UserDir to '%s'\n", server.config.UserDir)
	}
	if config.DataFileName == "" {
		server.config.DataFileName = "data.json"
	}

	server.store = store.Storer(&struct {
		store.UserStorer
		store.WatchlistStorer
	}{
		storefs.NewUserStore(server.config.UserDir, server.config.DataFileName, server.logger),
		storefs.NewWatchlistStore(server.ebidlocal, server.logger),
	})

	server.registerHTTPHandlers()

	return &server
}

type Server struct {
	logger    *log.Logger
	addr      *url.URL
	ebidlocal *ebidlocal.Ebidlocal
	store     store.Storer
	config    Config
}

func (s *Server) Run() {
	s.logger.Info.Printf("Listening on %s\n", s.addr.String())
	s.logger.Error.Fatal(http.ListenAndServe(s.addr.String(), nil))
}

func (s *Server) registerHTTPHandlers() {
	r := mux.NewRouter()

	s.registerUserRoutes(r.PathPrefix("/user").Subrouter())
	s.registerWatchlistRoutes(r.PathPrefix("/watchlist").Subrouter())

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static")))

	loggedRouter := handlers.RecoveryHandler()(handlers.LoggingHandler(os.Stdout, r))
	http.Handle("/", loggedRouter)
}

func (s *Server) registerUserRoutes(router *mux.Router) *mux.Router {
	router.Path("/{userID}/watchlist").Methods("POST").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.createUserWatchlistHandlerFunc), "application/json")).Name("createWatchlist")

	router.Path("/{userID}/watchlist/{listID}").Methods("GET").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		userID := params["userID"]
		listID := params["listID"]

		d := fmt.Sprintf("./web/watchlists/%s/", listID)
		rm := fmt.Sprintf("/user/%s/watchlist/%s", userID, listID)
		http.StripPrefix(rm, http.FileServer(http.Dir(d))).ServeHTTP(w, r)
	})).Name("getUserWatchlist")

	router.Methods("POST").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.createUserHandlerFunc), "application/json"))

	router.PathPrefix("/{userID}/data.json").Handler(http.StripPrefix("/user", http.FileServer(http.Dir("./web/user")))).Name("userData")

	router.PathPrefix("/{userID}/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := mux.Vars(r)["userID"]
		d := fmt.Sprintf("./web/user/%s/static", userID)
		rm := fmt.Sprintf("/user/%s/", userID)
		http.StripPrefix(rm, http.FileServer(http.Dir(d))).ServeHTTP(w, r)
	})).Name("userDir")

	return router
}

func (s *Server) registerWatchlistRoutes(router *mux.Router) *mux.Router {
	router.Methods("GET").Handler(http.StripPrefix("/watchlist", http.FileServer(http.Dir("./web/watchlists"))))
	return router
}

func (s *Server) createUserHandlerFunc(w http.ResponseWriter, r *http.Request) {
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

	user = NewUser(user.Name)
	if err = s.store.SaveUser(&user); err != nil {
		defer s.store.DeleteUser(user.ID)
		respondError(w, http.StatusInternalServerError, "User not created")
		s.logger.Error.Printf("Failed to create user %s", user.ID)
		return
	}
	if _, err = s.createUserSpace(&user); err != nil {
		respondError(w, http.StatusInternalServerError, "User not created")
		s.logger.Error.Printf("Failed to create user %s", user.ID)
		return
	}
	s.logger.Info.Printf("User created '%s'\n", user.ID)

	w.Header().Set("Location", fmt.Sprintf("/user/%s/", url.PathEscape(user.ID)))
	respondJSON(w, http.StatusCreated, user)
}

func (s *Server) createUserWatchlistHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var wl Watchlist

	userID := mux.Vars(r)["userID"]

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&wl); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !wl.IsValid() {
		respondError(w, http.StatusBadRequest, "User watchlist is required.")
		return
	}

	listID, err := s.addUserWatchlist(userID, &wl)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create watch list")
		return
	}

	s.logger.Info.Printf("Create watch list called '%s'", wl.Name)
	w.Header().Set("Location", fmt.Sprintf("/user/%s/watchlist/%s", url.PathEscape(userID), url.PathEscape(listID)))
	respondJSON(w, http.StatusCreated, struct {
		WatchlistID string `json:"watchlistID"`
	}{WatchlistID: listID})
}

func (s *Server) createUserSpace(u *User) (string, error) {
	var userDir string = filepath.Join(s.config.UserDir, u.ID)

	if err := os.MkdirAll(userDir, 0775); err != nil {
		s.logger.Error.Println(err)
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(userDir, "index.html"), []byte("<html><body>Nothing here yet."), 0644); err != nil {
		return "", err
	}

	absContentPath, err := filepath.Abs(s.config.ContentPath)
	if err != nil {
		return "", err
	}
	err = os.Symlink(filepath.Join(absContentPath, "template"), filepath.Join(userDir, "static"))
	if err != nil {
		return "", err
	}

	return userDir, nil
}

//addUserWatchlist add a watch list to a user's group of watch lists.
func (s *Server) addUserWatchlist(userID string, list *Watchlist) (string, error) {
	user, err := s.store.LoadUser(userID)
	if err != nil {
		return "", err
	}
	if err := s.store.SaveWatchlist(list); err != nil {
		return "", err
	}

	listID := watchlist.Watchlist(list.List).ID()
	user.Watchlists[list.Name] = listID
	return listID, s.store.SaveUser(user)
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
