package ebidlocal

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
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
	if config.ScanIncrement == 0 {
		config.ScanIncrement = 1
	}

	t, err := template.New("template.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
	}).ParseFiles(filepath.Join("./", "assets", "templates", "template.html.tmpl"))
	if err != nil {
		log.New("Ebidlocal.New").Error.Fatal(err)
	}

	return &Ebidlocal{
		config:   config,
		logger:   log.New("Ebidlocal"),
		template: t,
	}
}

//Ebidlocal data for ebidlocal app
type Ebidlocal struct {
	config   Config
	logger   *log.Logger
	template *template.Template
}

//Scan kick off directory scanner which keeps watchlists up-to-date.
func (e *Ebidlocal) Scan(done <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(e.config.ScanIncrement) * time.Minute)
	watchlistDir := filepath.Join(e.config.ContentPath, "web", "watchlists")
	e.logger.Info.Printf("Scanning '%s' at interval '%d'", watchlistDir, e.config.ScanIncrement)
	for {
		select {
		case <-ticker.C:
			err := filepath.Walk(watchlistDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
					return err
				}
				if info.Name() == "data.json" {
					fmt.Printf("Found file: %q\n", path)
					kw, err := e.loadWatchlist(path)
					if err != nil {
						e.logger.Error.Println(err)
						return nil
					}
					if file, err := os.Create(filepath.Join(filepath.Dir(path), "index.html")); err == nil {
						e.updateWatchlist(kw, file)
						file.Close()
					} else {
						e.logger.Error.Println(err)
					}
				}

				return nil
			})
			if err != nil {
				fmt.Printf("error walking the path %q: %v\n", watchlistDir, err)
				return
			}
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
func (e *Ebidlocal) AddUserWatchlist(userID string, watchlistName string, list Watchlist) (string, error) {
	user, err := e.loadUser(userID)
	if err != nil {
		return "", err
	}
	if err := e.addWatchlist(list); err != nil {
		return "", err
	}

	listID := list.ID()
	user.Watchlists[watchlistName] = listID
	return listID, e.saveUser(user)
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
	file, err := json.Marshal(u)
	if err != nil {
		e.logger.Error.Println(err)
		return err
	}
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
	var watchlistDir = filepath.Join(e.config.ContentPath, "web", "watchlists", list.ID())

	if _, err := os.Stat(watchlistDir); os.IsExist(err) {
		e.logger.Info.Println("Watch list already exists.")
		return nil
	}

	e.logger.Info.Printf("Creating watchlist.", watchlistDir)
	if err := os.MkdirAll(watchlistDir, 0775); err != nil {
		e.logger.Error.Println(err)
		return err
	}

	file, err := json.Marshal(list)
	if err != nil {
		e.logger.Error.Println(err)
		return err
	}

	return ioutil.WriteFile(filepath.Join(watchlistDir, "data.json"), file, 0644)
}

func (e *Ebidlocal) loadWatchlist(filePath string) (ebidlocal.Keywords, error) {
	var keywords ebidlocal.Keywords = make([]string, 0)

	jsonFile, err := os.Open(filePath)
	if err != nil {
		e.logger.Error.Println(err)
		return keywords, err
	}
	dec := json.NewDecoder(jsonFile)
	defer jsonFile.Close()
	if err := dec.Decode(&keywords); err != nil {
		e.logger.Error.Println(err)
		return keywords, err
	}
	e.logger.Info.Printf("Watch list found '%v'", keywords)
	return keywords, nil
}

func (e *Ebidlocal) updateWatchlist(keywords ebidlocal.Keywords, outFile io.Writer) error {
	if err := e.template.Execute(outFile, keywords.Search()); err != nil {
		e.logger.Error.Println(err)
		return err
	}
	return nil
}
