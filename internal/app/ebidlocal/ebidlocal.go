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
			e.logger.Info.Printf("Scanning '%s'", filepath.Join(e.config.ContentPath, e.config.UserDir))
		case <-done:
			ticker.Stop()
			return
		}
	}
}

//CreateUser create a new user, this also builds the user's workspace.
func (e *Ebidlocal) CreateUser(username string) (string, error) {
	u := NewUser(username)
	u.UserDir = e.createUserSpace(&u)
	e.saveUser(&u)

	return u.ID, nil
}

func (e *Ebidlocal) createUserSpace(u *User) string {
	var userDir string = filepath.Join(e.config.ContentPath, e.config.UserDir, u.ID)
	e.logger.Info.Printf("Creating user '%s' at '%s'\n", u.ID, userDir)
	os.MkdirAll(userDir, 0775)

	ioutil.WriteFile(filepath.Join(userDir, "index.html"), []byte("<html><body>"), 0644)

	cwd, err := os.Getwd()
	if err != nil {
		e.logger.Error.Fatalln(err)
	}
	e.logger.Info.Printf("Cwd '%s'\n", cwd)
	os.Symlink(filepath.Join(cwd, "template"), filepath.Join(userDir, "static"))

	return userDir
}

func (e *Ebidlocal) saveUser(u *User) {
	file, err := json.MarshalIndent(u, "", "    ")
	if err != nil {
		e.logger.Error.Fatalln(err)
	}
	e.logger.Info.Printf("Writing user data.", file)
	ioutil.WriteFile(filepath.Join(u.UserDir, e.config.DataFileName), file, 0644)
}

func (e *Ebidlocal) addWatchlist(u User, watchlistName string, list Watchlist) {
}
