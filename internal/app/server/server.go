package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/google/uuid"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/store"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/filter"
	ebidmodel "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/notify/email"
	stringutils "github.com/scirelli/auction-ebidlocal-search/internal/pkg/stringUtils"
)

func New(config Config, store store.Storer, logger log.Logger, searchExtractor SearchExtractor) *Server {
	var server = Server{
		config:          config,
		logger:          logger,
		store:           store,
		searchExtractor: searchExtractor,
	}

	t, err := template.New("verification.template.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
		"QueryEscape": func(queryString string) string {
			return url.QueryEscape(queryString)
		},
	}).ParseFiles(config.VerificationTemplateFile)
	if err != nil {
		logger.Fatal(err)
	}

	server.template = t
	server.addr = fmt.Sprintf("%s:%d", config.Address, config.Port)
	server.registerHTTPHandlers()

	return &server
}

type Server struct {
	logger          log.Logger
	addr            string
	store           store.Storer
	config          Config
	template        *template.Template
	searchExtractor SearchExtractor
}

func (s *Server) Run() {
	s.logger.Infof("Listening on %s\n", s.addr)
	s.logger.Fatal(http.ListenAndServe(s.addr, nil))
}

func (s *Server) registerHTTPHandlers() {
	r := mux.NewRouter()

	s.registerUserRoutes(r.PathPrefix("/user").Subrouter())
	s.registerWatchlistRoutes(r.PathPrefix("/watchlist").Subrouter())
	s.registerSearchRoutes(r.PathPrefix("/search").Subrouter())

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(filepath.Join(s.config.ContentPath, "/web/static"))))

	loggedRouter := handlers.RecoveryHandler()(handlers.LoggingHandler(os.Stdout, r))
	http.Handle("/", loggedRouter)
}

/*
Notes:
	Routes are tested in the order they were added to the router. If two routes match, the first one wins:
*/
func (s *Server) registerUserRoutes(router *mux.Router) *mux.Router {
	router.Path("/{userID}/watchlist").Methods("POST", "UPDATE").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.createUserWatchlistHandlerFunc), "application/json")).Name("createAndEditWatchlist")
	router.Path("/{userID}/watchlist").Methods("DELETE").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.deleteUserWatchlistHandlerFunc), "application/json")).Name("deleteWatchlist")

	router.Path("/{userID}/data.json").Methods("GET").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := mux.Vars(r)["userID"]
		user, err := s.store.LoadUser(r.Context(), userID)
		if err != nil {
			s.logger.Error(err)
			respondError(w, http.StatusBadRequest, "User id is required")
			return
		}

		//Never send sensitive data back to the clients. Do not send email, verification, or admin information.
		respondJSON(w, http.StatusCreated, struct {
			Name       string            `json:"name"`
			ID         string            `json:"id"`
			Verified   bool              `json:"verified"`
			Watchlists map[string]string `json:"watchlists"`
		}{
			Name:       user.Name,
			ID:         user.ID,
			Verified:   user.Verified,
			Watchlists: user.Watchlists,
		})
	})).Name("userData")

	router.PathPrefix("/{userID}/watchlist/{listID}/").Methods("GET").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		userID := params["userID"]
		listID := params["listID"]

		d := filepath.Join(s.config.WatchlistDir, listID)
		rm := fmt.Sprintf("/user/%s/watchlist/%s", userID, listID)
		http.StripPrefix(rm, http.FileServer(http.Dir(d))).ServeHTTP(w, r)
	})).Name("getUserWatchlist")

	router.Methods("POST").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.createUserHandlerFunc), "application/json")).Name("createUser")

	router.Path("/{userID}/verify/send").Methods("PUT").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.sendUserVerificationHandlerFunc), "application/json", "text/plain")).Name("sendUserVerification")
	router.Path("/{userID}/verify/{nonce}").Methods("UPDATE", "GET").Handler(handlers.ContentTypeHandler(http.HandlerFunc(s.verifyUserHandlerFunc), "application/json", "text/plain")).Name("verifyUser")
	router.Path("/{userID}/verify").Methods("GET").Handler(http.HandlerFunc(s.isVerifiedUserHandlerFunc)).Name("isVerifiedUser")

	router.PathPrefix("/{userID}/").Methods("GET").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := mux.Vars(r)["userID"]
		d := filepath.Join(s.config.UserDir, fmt.Sprintf("/%s/static", userID))
		rm := fmt.Sprintf("/user/%s/", userID)
		http.StripPrefix(rm, http.FileServer(http.Dir(d))).ServeHTTP(w, r)
	})).Name("userDir")

	return router
}

func (s *Server) registerWatchlistRoutes(router *mux.Router) *mux.Router {
	router.Methods("GET").Handler(http.StripPrefix("/watchlist", http.FileServer(http.Dir(s.config.WatchlistDir))))
	return router
}

func (s *Server) registerSearchRoutes(router *mux.Router) *mux.Router {
	router.Path("/").Methods("GET").Queries("q", "{q}").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//query := r.FormValue("q")
		var q []string
		var exists bool
		var results []ebidmodel.AuctionItem
		query := r.URL.Query()
		if q, exists = query["q"]; !exists || (exists && len(stringutils.FilterEmpty(q)) == 0) {
			respondError(w, http.StatusBadRequest, "Missing query value")
			return
		}
		for result := range ebidmodel.FilterAuctionItemChan(s.searchExtractor.Extract(s.searchExtractor.Search(stringiter.SliceStringIterator(q)))).Filter(ebidmodel.FilterFunc(filter.ByKeyword)) {
			s.logger.Info(result)
			results = append(results, result)
		}
		respondJSON(w, http.StatusOK, results)
	})).Name("Quick-Search")
	return router
}

func (s *Server) createUserHandlerFunc(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user model.User
	var err error
	var nonce uuid.UUID

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, "User data is required")
		return
	}
	if !user.IsValid() {
		respondError(w, http.StatusBadRequest, "User name and email are required.")
		return
	}

	tmp := model.NewUser(user.Name)
	tmp.Email = user.Email
	user = tmp

	if nonce, err = s.sendUserVerification(&user); err != nil {
		s.logger.Errorf("Fialed to send email verification for user (%s); '%s'", user.ID, err)
	}
	user.VerifyToken = nonce

	if _, err = s.store.SaveUser(r.Context(), &user); err != nil {
		defer s.store.DeleteUser(r.Context(), user.ID)
		respondError(w, http.StatusInternalServerError, "User not created")
		s.logger.Errorf("Failed to create user %s", user.ID)
		return
	}
	if _, err = s.createUserSpace(&user); err != nil {
		respondError(w, http.StatusInternalServerError, "User not created")
		s.logger.Errorf("Failed to create user %s", user.ID)
		return
	}
	s.logger.Infof("User created '%s'\n", user.ID)
	w.Header().Set("Location", fmt.Sprintf("/user/%s/", url.PathEscape(user.ID)))
	respondJSON(w, http.StatusCreated, user)
}

func (s *Server) createUserWatchlistHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var wl model.Watchlist

	userID := mux.Vars(r)["userID"]

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&wl); err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !wl.IsValid() {
		s.logger.Error("User watchlist is required.")
		respondError(w, http.StatusBadRequest, "User watchlist is required.")
		return
	}

	listID, err := s.addUserWatchlist(r.Context(), userID, &wl)
	if os.IsNotExist(err) {
		respondError(w, http.StatusNotFound, "Unknown User")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create watch list.")
		return
	}

	s.logger.Infof("Create watch list called '%s'", wl.Name)
	w.Header().Set("Location", fmt.Sprintf("/user/%s/watchlist/%s", url.PathEscape(userID), url.PathEscape(listID)))
	respondJSON(w, http.StatusCreated, struct {
		WatchlistID string `json:"watchlistID"`
	}{WatchlistID: listID})
}

func (s *Server) deleteUserWatchlistHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var wl model.Watchlist

	userID := mux.Vars(r)["userID"]

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&wl); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if wl.Name == "" {
		respondError(w, http.StatusBadRequest, "Watch list name is required")
		return
	}

	err := s.deleteUserWatchlist(r.Context(), userID, &wl)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete watch list")
		return
	}

	s.logger.Infof("Delete watch list called '%s'", wl.Name)
	respondJSON(w, http.StatusNoContent, "Deleted")
}

func (s *Server) sendUserVerificationHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var nonce uuid.UUID
	var userID string
	var err error
	var user *model.User

	userID = mux.Vars(r)["userID"]
	user, err = s.store.LoadUser(r.Context(), userID)
	if err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusNotFound, "User Not Found")
		return
	}

	if nonce, err = s.sendUserVerification(user); err != nil {
		s.logger.Error(err)
		switch interface{}(err).(type) {
		case VerifyTooEarlyError:
			respondError(w, http.StatusTooEarly, "Verification Already Sent")
		default:
			respondError(w, http.StatusInternalServerError, "Send Verification Failed")
		}
		return
	}

	user.VerifyToken = nonce
	user.LastVerified = time.Now()
	if _, err := s.store.SaveUser(r.Context(), user); err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusInternalServerError, "User Save Failed")
		return
	}

	respondJSON(w, http.StatusAccepted, struct {
		Send bool `json:"send"`
	}{
		Send: true,
	})
}

func (s *Server) verifyUserHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var userID string
	var nonce uuid.UUID
	var err error
	var user *model.User

	userID = mux.Vars(r)["userID"]
	user, err = s.store.LoadUser(r.Context(), userID)
	if err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusNotFound, "User Not Found")
		return
	}

	if nonce, err = uuid.Parse(mux.Vars(r)["nonce"]); err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusBadRequest, "Invalid Nonce")
		return
	}

	if err := s.verifyUser(user, nonce); err != nil {
		s.logger.Error(err)
		switch interface{}(err).(type) {
		case VerifyTokenExpiredError:
			respondError(w, http.StatusPreconditionFailed, "Token Expired")
		case VerifyBadTokenError:
			respondError(w, http.StatusUnprocessableEntity, "Bad Token")
		default:
			respondError(w, http.StatusInternalServerError, "Send Verification Failed")
		}
		return
	}

	user.Verified = true
	if _, err := s.store.SaveUser(r.Context(), user); err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusInternalServerError, "User Save Failed")
		return
	}

	urlParams := r.URL.Query()
	redirect := urlParams["redirect"][0]
	status := http.StatusOK
	if redirect != "" {
		redirect, err = url.QueryUnescape(redirect)
		w.Header().Set("Location", redirect)
		status = http.StatusSeeOther
	}

	respondJSON(w, status, struct {
		Verified bool `json:"verified"`
	}{
		Verified: true,
	})
}

func (s *Server) isVerifiedUserHandlerFunc(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	user, err := s.store.LoadUser(r.Context(), userID)
	if err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusNotFound, "User Not Found")
		return
	}

	if err := s.isVerifiedUser(user); err != nil {
		s.logger.Error(err)
		respondError(w, http.StatusUnauthorized, "Not Verified")
		return
	}

	respondJSON(w, http.StatusOK, struct {
		Verified bool `json:"verified"`
	}{
		Verified: true,
	})
}

//sendUserVerification sends a verification email to the user. First it checks the last verified timestamp to prevent spamming of verification emails. If interval has elapsed it will send a
// new varify url to the user.
func (s *Server) sendUserVerification(u *model.User) (uuid.UUID, error) {
	var emailBody bytes.Buffer
	nonce := uuid.New()

	if err := s.createVerificationEmail(u, nonce, &emailBody); err != nil {
		return uuid.UUID{}, err
	}
	if err := email.NewEmail(
		[]string{u.Email},
		fmt.Sprintf("Ebidlocal Watch List Email Verification  for'%s'", u.Name),
		string(emailBody.String()),
	).Send(); err != nil {
		return uuid.UUID{}, err
	}

	return nonce, nil
}

func (s *Server) createVerificationEmail(u *model.User, nonce uuid.UUID, out io.Writer) error {
	var base, pathUrl *url.URL
	var err error

	base, err = url.Parse(s.config.ServerUrl)
	if err != nil {
		return err
	}
	pathUrl, err = url.Parse(path.Join("user", u.ID, "verify", nonce.String()))
	if err != nil {
		return err
	}

	if err := s.template.Execute(out, struct {
		VerificationLink string
		WatchlistLink    string
	}{
		VerificationLink: base.ResolveReference(pathUrl).String(),
		WatchlistLink:    fmt.Sprintf("%s/viewwatchlists.html?id=%s", s.config.UiUrl, url.QueryEscape(u.ID)),
	}); err != nil {
		return err
	}

	return nil
}

//verifyUser verifies the user, if user is verified a nil error is returned, otherwise an error is returned.
func (s *Server) verifyUser(user *model.User, nonce uuid.UUID) error {
	if user.Verified {
		return nil
	}

	if !time.Now().Before(user.LastVerified.Add(s.config.VerificationWindowMinutes * time.Minute)) {
		return &VerifyTokenExpiredError{}
	}

	if user.VerifyToken != nonce {
		return &VerifyBadTokenError{}
	}

	return nil
}

func (s *Server) isVerifiedUser(user *model.User) error {
	if !user.Verified {
		return &VerifyNotVerifiedError{}
	}
	return nil
}

func (s *Server) createUserSpace(u *model.User) (string, error) {
	var userDir string = filepath.Join(s.config.UserDir, u.ID)

	if err := os.MkdirAll(userDir, 0775); err != nil {
		s.logger.Error(err)
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(userDir, "index.html"), []byte("<html><body>Nothing here yet."), 0644); err != nil {
		return "", err
	}

	absContentPath, err := filepath.Abs(s.config.ContentPath)
	if err != nil {
		return "", err
	}
	err = os.Symlink(filepath.Join(absContentPath, "web", "static"), filepath.Join(userDir, "static"))
	if err != nil {
		return "", err
	}

	return userDir, nil
}

//addUserWatchlist add a watch list to a user's group of watch lists.
func (s *Server) addUserWatchlist(ctx context.Context, userID string, list *model.Watchlist) (listID string, err error) {
	user, err := s.store.LoadUser(ctx, userID)
	if err != nil {
		return "", err
	}

	if listID, err = s.store.SaveWatchlist(ctx, list); err != nil {
		return "", err
	}

	user.Watchlists[list.Name] = listID
	if _, err = s.store.SaveUser(ctx, user); err != nil {
		//Don't care about the watch list that was saved, since watch lists can be shared among users. No reason to delete it.
		return "", err
	}

	return listID, nil
}

//deleteUserWatchlist delete a watch list from a user's group of watch lists.
func (s *Server) deleteUserWatchlist(ctx context.Context, userID string, list *model.Watchlist) error {
	user, err := s.store.LoadUser(ctx, userID)
	if err != nil {
		return err
	}

	delete(user.Watchlists, list.Name)
	if _, err = s.store.SaveUser(ctx, user); err != nil {
		return err
	}

	return nil
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
