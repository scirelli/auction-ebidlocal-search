package scanner

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"

	search "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//New constructor for scanner app. This app creates new watch lists on disk and has a scanner to keep them up-to-date.
func New(config Config) *Scanner {
	var logger = log.New("Scanner.New")

	t, err := template.New("template.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
	}).ParseFiles(filepath.Join("./", "assets", "templates", "template.html.tmpl"))

	if err != nil {
		logger.Error.Fatal(err)
	}

	return &Scanner{
		config:          config,
		logger:          logger,
		template:        t,
		watchlists:      make(chan string),
		auctionSearcher: search.AuctionSearchFunc(search.SearchAuctions),
	}
}

//Scanner data for scanner bidlocal app
type Scanner struct {
	config          Config
	logger          *log.Logger
	template        *template.Template
	watchlists      chan string
	openAuctions    stringiter.Iterable
	auctionSearcher search.AuctionSearcher
	store           store.Storer
}

func (e *Scanner) SetOpenAuctions(openAuctions stringiter.Iterable) *Scanner {
	e.openAuctions = openAuctions
	return e
}

//Scan kick off directory scanner which keeps watchlists up-to-date.
func (e *Scanner) Scan(ctx context.Context) {
	go func() {
		for path := range e.findWatchlists(ctx) {
			e.watchlists <- path
		}
	}()

	//TODO: Fix this to re-queue failed updates.
	//TODO: Fix to make failed requests back off and eventually die.
	//TODO: Fix to rate limit requests.
	//TODO: Email to. With verification.
	e.batchUpdateWatchlists(10 * time.Second)
}

//EnqueueWatchlist takes a watch list, builds the path to the watch list data file, and puts it on the watch list queue.
func (e *Scanner) EnqueueWatchlist(list watchlist.Watchlist) {
	watchlistFile := filepath.Join(e.config.WatchlistDir, list.ID(), "data.json")
	go func() {
		e.watchlists <- watchlistFile
	}()
}

// findWatchlists walk the watch list directory on an internval.
// returns a chan of paths to the watch list data file.
func (e *Scanner) findWatchlists(ctx context.Context) <-chan string {
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
			case <-ctx.Done():
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

//batchUpdateWatchlists Batch update watch lists. Makes four requests at a time.
func (e *Scanner) batchUpdateWatchlists(runInterval time.Duration) {
	for path := range e.watchlists {
		var wg sync.WaitGroup
		wg.Add(4)
		startTime := time.Now()
		go func() {
			defer wg.Done()
			if err := e.updateWatchlistResults(path); err != nil {
				return
			}
			e.notifyOnChange(path)
		}()
		go func() {
			defer wg.Done()
			path := <-e.watchlists
			if err := e.updateWatchlistResults(path); err != nil {
				return
			}
			e.notifyOnChange(path)
		}()
		go func() {
			defer wg.Done()
			path := <-e.watchlists
			if err := e.updateWatchlistResults(path); err != nil {
				return
			}
			e.notifyOnChange(path)
		}()
		go func() {
			defer wg.Done()
			path := <-e.watchlists
			if err := e.updateWatchlistResults(path); err != nil {
				return
			}
			e.notifyOnChange(path)
		}()
		wg.Wait()
		if elaspsedTime := time.Since(startTime); elaspsedTime < runInterval {
			time.Sleep(runInterval - elaspsedTime)
		}
	}
}

func (e *Scanner) notifyOnChange(watchlistFilePath string) error {
	var id string
	var err error
	if id, err = getResultID(e.watchlistFileFromPath(watchlistFilePath)); err != nil {
		e.logger.Error.Println(err)
		return err
	}
	e.logger.Info.Println(id)
	return nil
}

//updateWathclistResults loads a watch list, makes a request to ebid for new search results.
func (e *Scanner) updateWatchlistResults(watchlistFilePath string) error {
	watchlist, err := e.loadWatchlist(watchlistFilePath)
	if err != nil {
		e.logger.Error.Println(err)
		return err
	}

	if file, err := os.Create(e.watchlistFileFromPath(watchlistFilePath)); err == nil {
		defer file.Close()
		if err := e.template.Execute(file, e.auctionSearcher.Search(stringiter.SliceStringIterator(watchlist), e.openAuctions)); err != nil {
			e.logger.Error.Println(err)
			return err
		}
	} else {
		e.logger.Error.Println(err)
		return err
	}

	return nil
}

func (e *Scanner) watchlistFileFromPath(watchlistFilePath string) string {
	return filepath.Join(filepath.Dir(watchlistFilePath), "index.html")
}

//loadWatchlist loads a watch list from file.
func (e *Scanner) loadWatchlist(filePath string) (watchlist.Watchlist, error) {
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

func getResultID(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
