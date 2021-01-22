package ebidlocal

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal/watchlist"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//New constructor for ebidlocal app.
func New(config Config) *Ebidlocal {
	var logger = log.New("Ebidlocal.New")

	if config.ContentPath == "" {
		config.ContentPath = "."
		logger.Info.Printf("Defaulting content path dir to '%s'\n", config.ContentPath)
	}
	if config.TemplateDir == "" {
		config.TemplateDir = "/template"
		logger.Info.Printf("Defaulting template dir to '%s'\n", config.TemplateDir)
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Info.Printf("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}
	if config.ScanInterval == 0 {
		config.ScanInterval = 1
		logger.Info.Printf("Defaulting scan interval to '%d'\n", config.ScanInterval)
	}

	t, err := template.New("template.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
	}).ParseFiles(filepath.Join("./", "assets", "templates", "template.html.tmpl"))
	if err != nil {
		logger.Error.Fatal(err)
	}

	return &Ebidlocal{
		config:     config,
		logger:     logger,
		template:   t,
		watchlists: make(chan string),
	}
}

//Ebidlocal data for ebidlocal app
type Ebidlocal struct {
	config     Config
	logger     *log.Logger
	template   *template.Template
	watchlists chan string
}

//Scan kick off directory scanner which keeps watchlists up-to-date.
func (e *Ebidlocal) Scan(done <-chan struct{}) {
	go func() {
		for path := range e.findWatchlists(done) {
			e.watchlists <- path
		}
	}()

	go func() {
		//TODO: Fix this to requeue failed updates.
		//TODO: Fix to make failed requets back off and eventually die.
		//TODO: Fix to rate limit requets.
		//TODO: Email to. With verification.
		for path := range e.watchlists {
			e.updateWathclist(path)
		}
	}()
}

func (e *Ebidlocal) findWatchlists(done <-chan struct{}) <-chan string {
	timeBetweenRuns := time.Duration(e.config.ScanInterval) * time.Second
	watchlistDir := e.config.WatchlistDir
	foundWatchlists := make(chan string)

	e.logger.Info.Printf("Scanning '%s' at interval '%d' minutes", watchlistDir, e.config.ScanInterval)
	go func() {
		defer close(foundWatchlists)
		for {
			startTime := time.Now()

			if err := filepath.Walk(watchlistDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					e.logger.Info.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
					return err
				}
				if info.Name() == "data.json" {
					e.logger.Info.Printf("Found file: %q\n", path)
					foundWatchlists <- path
				}

				return nil
			}); err != nil {
				e.logger.Error.Printf("Error walking the path %q: %v\n", watchlistDir, err)
			}

			select {
			case <-done:
				return
			default:
			}
			if elaspsedTime := time.Since(startTime); elaspsedTime < timeBetweenRuns {
				time.Sleep(timeBetweenRuns - elaspsedTime)
			}
		}
	}()

	return foundWatchlists
}

//AddWatchlist saves a watchlist to disk. Skips saving if it already exists.
func (e *Ebidlocal) AddWatchlist(list watchlist.Watchlist) error {
	var watchlistDir = filepath.Join(e.config.WatchlistDir, list.ID())

	e.logger.Info.Printf("Checking for '%s'\n", watchlistDir)
	if _, err := os.Stat(watchlistDir); os.IsExist(err) {
		e.logger.Info.Println("Watch list already exists.")
		return nil
	}

	e.logger.Info.Printf("Creating watchlist. '%s'", watchlistDir)
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

func (e *Ebidlocal) updateWathclist(watchListFilePath string) error {
	watchlist, err := e.loadWatchlist(watchListFilePath)
	if err != nil {
		e.logger.Error.Println(err)
		return err
	}

	if file, err := os.Create(filepath.Join(filepath.Dir(watchListFilePath), "index.html")); err == nil {
		defer file.Close()
		if err := e.template.Execute(file, ebidlocal.Keywords(watchlist).Search()); err != nil {
			e.logger.Error.Println(err)
			return err
		}
	} else {
		e.logger.Error.Println(err)
		return err
	}

	return nil
}

func (e *Ebidlocal) loadWatchlist(filePath string) (watchlist.Watchlist, error) {
	var watchlist watchlist.Watchlist = make([]string, 0)

	jsonFile, err := os.Open(filePath)
	if err != nil {
		e.logger.Error.Println(err)
		return watchlist, err
	}
	defer jsonFile.Close()
	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&watchlist); err != nil {
		e.logger.Error.Println(err)
		return watchlist, err
	}
	e.logger.Info.Printf("Watch list found '%v'", watchlist)
	return watchlist, nil
}
