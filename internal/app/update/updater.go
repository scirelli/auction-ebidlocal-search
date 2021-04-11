package update

import (
	"context"
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	search "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/publish"
)

type Updater interface {
	Update(watchlistPath <-chan string) error
}

//New constructor for updater app. This app creates new watch lists on disk and has a updater to keep them up-to-date.
func New(ctx context.Context, watchlistStore store.Storer, config Config) *Update {
	var logger = log.New("Update")

	t, err := template.New("template.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
	}).ParseFiles(filepath.Join("./", "assets", "templates", "template.html.tmpl"))

	if err != nil {
		logger.Error.Fatal(err)
	}

	return &Update{
		config:          config,
		logger:          logger,
		template:        t,
		auctionSearcher: search.AuctionSearchFunc(search.SearchAuctions),
		watchlistIDs:    make(chan string, config.BatchSize),
		ctx:             ctx,
		store:           watchlistStore,
		changePublsr:    publish.NewStringChange(),
	}
}

type Update struct {
	config          Config
	logger          *log.Logger
	template        *template.Template
	openAuctions    stringiter.Iterable
	auctionSearcher search.AuctionSearcher
	store           store.Storer
	watchlistIDs    chan string
	ctx             context.Context
	changePublsr    publish.StringPublisher
}

func (e *Update) SetOpenAuctions(openAuctions stringiter.Iterable) *Update {
	e.openAuctions = openAuctions
	return e
}

func (e *Update) Update(watchlistFilePaths <-chan string) error {
	go e.batchUpdateWatchlists()
	for path := range watchlistFilePaths {
		if err := e.EnqueueWatchlistPath(path); err != nil {
			e.logger.Error.Printf("%s", err)
			return err
		}
	}

	return nil
}

func (e *Update) SubscribeForChange() (<-chan string, func() error) {
	return e.changePublsr.Subscribe()
}

//EnqueueWatchlistPath takes a path to a watch list converts it to an ID and puts it on the watch list update queue.
func (e *Update) EnqueueWatchlistPath(watchlistFilePath string) error {
	var id string
	id = e.watchlistIDFromPath(filepath.Dir(watchlistFilePath))
	e.logger.Info.Printf("Calling Enqueueing watchlist id: '%s'; path: '%s'", id, watchlistFilePath)
	e.EnqueueWatchlistID(id)
	return nil
}

//EnqueueWatchlistID takes a watch list id and puts it on the watch list update queue.
func (e *Update) EnqueueWatchlistID(listID string) {
	go func(listID string) {
		e.watchlistIDs <- listID
	}(listID)
}

//TODO: Fix this to re-queue failed updates.
//TODO: Fix to make failed requests back off and eventually die.
//TODO: Email to. With verification.
//batchUpdateWatchlists Batch update watch lists. Makes x requests at a time.
func (e *Update) batchUpdateWatchlists() {
	e.logger.Info.Println("Batch Updates started.")
	var runInterval time.Duration = time.Duration(e.config.RunIntervalSeconds) * time.Second
	for {
		var wg sync.WaitGroup
		startTime := time.Now()
		for i := uint64(0); i < e.config.BatchSize; i++ {
			id := <-e.watchlistIDs
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				if err := e.updateWatchlistResults(id); err != nil {
					return
				}
				e.notifyOnChange(id)
			}(id)
		}
		wg.Wait()
		if elaspsedTime := time.Since(startTime); elaspsedTime < runInterval {
			time.Sleep(runInterval - elaspsedTime)
		}
	}
}

//updateWathclistResults loads a watch list, makes a request to ebid for new search results.
func (e *Update) updateWatchlistResults(id string) error {
	e.logger.Info.Printf("Updating watch list id: '%s'", id)

	watchlist, err := e.store.LoadWatchlist(context.Background(), id)
	if err != nil {
		e.logger.Error.Println(err)
		return err
	}

	if file, err := os.Create(e.watchlistDataFilePathFromID(id)); err == nil {
		defer file.Close()
		if err := e.template.Execute(file, e.auctionSearcher.Search(stringiter.SliceStringIterator(watchlist), e.openAuctions)); err != nil {
			e.logger.Error.Println(err)
			return err
		}
		e.logger.Info.Println("Generate file ID")
	} else {
		e.logger.Error.Println(err)
		return err
	}

	return nil
}

func (e *Update) notifyOnChange(watchlistID string) {
	var resultID string
	var err error

	if resultID, err = getResultID(e.watchlistDataFilePathFromID(watchlistID)); err != nil {
		e.logger.Error.Printf("Failed to check for changes. '%s'", err)
		return
	}

	if file, err := os.OpenFile(e.watchlistHashFilePathFromID(watchlistID), os.O_RDWR|os.O_CREATE, 0644); err == nil {
		defer file.Close()
		shaLength := 40
		buf := make([]byte, shaLength)
		if cnt, err := file.Read(buf); err != nil && err != io.EOF {
			e.logger.Error.Printf("Bytes read: '%d'; '%s'", cnt, err)
			return
		}
		if resultID != string(buf) {
			e.logger.Info.Printf("There was a change '%s' != '%s'", resultID, string(buf))
			if cnt, err := file.WriteAt([]byte(resultID), 0); err != nil || cnt != len(resultID) {
				e.logger.Error.Printf("Bytes written: '%d'; '%s'", cnt, err)
				return
			}
			e.changePublsr.Publish(watchlistID)
		}
	} else {
		e.logger.Error.Println(err)
		return
	}
}

func (e *Update) watchlistPathFromID(watchlistID string) string {
	return filepath.Join(e.config.WatchlistDir, watchlistID)
}

func (e *Update) watchlistDataFilePathFromID(watchlistID string) string {
	return filepath.Join(e.config.WatchlistDir, watchlistID, "index.html")
}

func (e *Update) watchlistHashFilePathFromID(watchlistID string) string {
	return filepath.Join(e.config.WatchlistDir, watchlistID, "hash")
}

func (e *Update) watchlistIDFromPath(watchlistFilePath string) string {
	_, file := filepath.Split(watchlistFilePath)
	e.logger.Info.Printf("Getting id from path '%s' - '%s'", watchlistFilePath, file)
	return file
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
