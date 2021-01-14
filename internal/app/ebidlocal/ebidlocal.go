package ebidlocal

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//New constructor for ebidlocal app.
func New(config Config) *Ebidlocal {
	if config.UserDir == "" {
		config.UserDir = "/web/user/"
	}
	if config.TemplateDir == "" {
		config.TemplateDir = "/template"
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}
	if config.WatchlistDirName == "" {
		config.WatchlistDirName = "watchlists"
	}

	return &Ebidlocal{
		config: config,
		logger: log.New("Ebidlocal"),
	}
}

//Ebidlocal data for ebidlocal app
type Ebidlocal struct {
	config Config
	logger *log.Logger
}

//Scan kick off directory scanner which keeps watchlists up-to-date.
func (e *Ebidlocal) Scan(done <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			e.logger.Info.Printf("Scanning '%s'", filepath.Join(e.config.ContentPath, e.config.UserDir, "watchlists"))
		case <-done:
			ticker.Stop()
			return
		}
	}
}

//CreateUser create a new user, this also builds the user's workspace.
func (e *Ebidlocal) CreateUser(username string) (string, error) {
	var err error
	u := NewUser(username)
	u.UserDir, err = e.createUserSpace(&u)
	if err != nil {
		return "", err
	}
	err = e.saveUser(&u)
	if err != nil {
		return "", err
	}

	return u.ID, nil
}

//AddUserWatchlist add a watch list to a user's group of watch lists.
func (e *Ebidlocal) AddUserWatchlist(userID string, watchlistName string, list Watchlist) error {
	user, err := e.loadUser(userID)
	if err != nil {
		return err
	}
	user.Watchlists[watchlistName] = list
	return e.saveUser(user)
}

func (e *Ebidlocal) createUserSpace(u *User) (string, error) {
	var userDir string = filepath.Join(e.config.ContentPath, e.config.UserDir, u.ID)
	e.logger.Info.Printf("Creating user '%s' at '%s'\n", u.ID, userDir)
	os.MkdirAll(userDir, 0775)

	if err := ioutil.WriteFile(filepath.Join(userDir, "index.html"), []byte("<html><body>"), 0644); err != nil {
		return "", err
	}

	absContentPath, err := filepath.Abs(e.config.ContentPath)
	if err != nil {
		return "", err
	}
	err = os.Symlink(filepath.Join(absContentPath, "template"), filepath.Join(userDir, "static"))
	if err != nil {
		return "", err
	}

	return userDir, nil
}

func (e *Ebidlocal) saveUser(u *User) error {
	file, err := json.MarshalIndent(u, "", "    ")
	if err != nil {
		e.logger.Error.Println(err)
		return err
	}
	e.logger.Info.Printf("Writing user data.", file)
	return ioutil.WriteFile(filepath.Join(u.UserDir, e.config.DataFileName), file, 0644)
}

func (e *Ebidlocal) loadUser(userID string) (*User, error) {
	var userDataFile string = filepath.Join(e.config.ContentPath, e.config.UserDir, userID, e.config.DataFileName)

	if _, err := os.Stat(userDataFile); os.IsNotExist(err) {
		e.logger.Info.Println("User does not exist")
		return nil, err
	}

	var usr User
	jsonFile, err := os.Open(userDataFile)
	if err != nil {
		e.logger.Error.Println(err)
		return nil, err
	}
	dec := json.NewDecoder(jsonFile)
	defer jsonFile.Close()
	if err := dec.Decode(&usr); err != nil {
		e.logger.Error.Println(err)
		return nil, err
	}

	return &usr, nil
}

func (e *Ebidlocal) addWatchlist(list Watchlist) error {
	var watchlistDir = filepath.Join(e.config.ContentPath, "watchlists", list.ID())

	if _, err := os.Stat(watchlistDir); os.IsExist(err) {
		e.logger.Info.Println("Watch list already exists.")
		return nil
	}

	if err := os.MkdirAll(watchlistDir, 0775); err != nil {
		e.logger.Error.Println(err)
		return err
	}

	file, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		e.logger.Error.Println(err)
		return err
	}

	return ioutil.WriteFile(filepath.Join(watchlistDir, "data.json"), file, 0644)
}
