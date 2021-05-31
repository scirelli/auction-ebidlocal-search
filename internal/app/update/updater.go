package update

import (
	"context"
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

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
	var logger = log.New("Update", log.DEFAULT_LOG_LEVEL)

	t, err := template.New("template.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
	}).ParseFiles(filepath.Join("./", "assets", "templates", "template.html.tmpl"))

	if err != nil {
		logger.Fatal(err)
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

//Update data
type Update struct {
	config          Config
	logger          log.Logger
	template        *template.Template
	openAuctions    stringiter.Iterable
	auctionSearcher search.AuctionSearcher
	store           store.Storer
	watchlistIDs    chan string
	ctx             context.Context
	changePublsr    publish.StringPublisher
}

//SetOpenAuctions sets the list of open auctions to search.
func (u *Update) SetOpenAuctions(openAuctions stringiter.Iterable) *Update {
	u.openAuctions = openAuctions
	return u
}

//Update starts the batch update of watch lists, reading from watchlistFilePaths channel and enqueuing them to be updated.
func (u *Update) Update(watchlistFilePaths <-chan string) error {
	go u.batchUpdateWatchlists()
	for path := range watchlistFilePaths {
		if err := u.EnqueueWatchlistPath(path); err != nil {
			u.logger.Errorf("%s", err)
			return err
		}
	}

	return nil
}

//SubscribeForChange returns a channel that can be monitored for changes, it also returns a function to call unsubscribe the channel.
func (u *Update) SubscribeForChange() (<-chan string, func() error) {
	return u.changePublsr.Subscribe()
}

//EnqueueWatchlistPath takes a path to a watch list converts it to an ID and puts it on the watch list update queue.
func (u *Update) EnqueueWatchlistPath(watchlistFilePath string) error {
	var id string
	id = u.watchlistIDFromPath(filepath.Dir(watchlistFilePath))
	u.logger.Infof("Calling Enqueueing watchlist id: '%s'; path: '%s'", id, watchlistFilePath)
	u.EnqueueWatchlistID(id)
	return nil
}

//EnqueueWatchlistID takes a watch list id and puts it on the watch list update queue.
func (u *Update) EnqueueWatchlistID(listID string) {
	go func(listID string) {
		u.watchlistIDs <- listID
	}(listID)
}

//batchUpdateWatchlists Batch update watch lists. Makes x requests at a time.
func (u *Update) batchUpdateWatchlists() {
	u.logger.Info("Batch Updates started.")
	var runInterval time.Duration = time.Duration(u.config.RunIntervalSeconds) * time.Second
	for {
		var wg sync.WaitGroup
		startTime := time.Now()
		for i := uint64(0); i < u.config.BatchSize; i++ {
			id := <-u.watchlistIDs
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				if err := u.updateWatchlistResults(id); err != nil {
					return
				}
				u.notifyOnChange(id)
			}(id)
		}
		wg.Wait()
		if elaspsedTime := time.Since(startTime); elaspsedTime < runInterval {
			time.Sleep(runInterval - elaspsedTime)
		}
	}
}

//updateWathclistResults loads a watch list, makes a request to ebid for new search results.
func (u *Update) updateWatchlistResults(id string) error {
	u.logger.Infof("Updating watch list id: '%s'", id)

	watchlist, err := u.store.LoadWatchlist(context.Background(), id)
	if err != nil {
		u.logger.Error(err)
		return err
	}

	if file, err := os.Create(u.watchlistDataFilePathFromID(id)); err == nil {
		defer file.Close()
		if err := u.template.Execute(file, struct {
			Rows          chan string
			WatchlistLink string
			WatchlistName string
		}{
			Rows:          u.auctionSearcher.Search(stringiter.SliceStringIterator(watchlist), u.openAuctions),
			WatchlistLink: u.config.ServerUrl + "/watchlist/" + id,
			WatchlistName: "<!--{{watchlistName}}-->",
		}); err != nil {
			u.logger.Error(err)
			return err
		}
		u.logger.Info("Generate file ID")
	} else {
		u.logger.Error(err)
		return err
	}

	return nil
}

func (u *Update) notifyOnChange(watchlistID string) {
	var resultID string
	var err error

	if resultID, err = u.getResultID(u.watchlistDataFilePathFromID(watchlistID)); err != nil {
		u.logger.Errorf("Failed to check for changes. '%s'", err)
		return
	}

	if file, err := os.OpenFile(u.watchlistHashFilePathFromID(watchlistID), os.O_RDWR|os.O_CREATE, 0644); err == nil {
		defer file.Close()
		shaLength := 40
		buf := make([]byte, shaLength)
		if cnt, err := file.Read(buf); err != nil && err != io.EOF {
			u.logger.Errorf("Bytes read: '%d'; '%s'", cnt, err)
			return
		}
		if resultID != string(buf) {
			u.logger.Infof("There was a change '%s' != '%s'", resultID, string(buf))
			if cnt, err := file.WriteAt([]byte(resultID), 0); err != nil || cnt != len(resultID) {
				u.logger.Errorf("Bytes written: '%d'; '%s'", cnt, err)
				return
			}
			u.changePublsr.Publish(watchlistID)
		}
	} else {
		u.logger.Error(err)
		return
	}
}

func (u *Update) watchlistPathFromID(watchlistID string) string {
	return filepath.Join(u.config.WatchlistDir, watchlistID)
}

func (u *Update) watchlistDataFilePathFromID(watchlistID string) string {
	return filepath.Join(u.config.WatchlistDir, watchlistID, "index.html")
}

func (u *Update) watchlistHashFilePathFromID(watchlistID string) string {
	return filepath.Join(u.config.WatchlistDir, watchlistID, "hash")
}

func (u *Update) watchlistIDFromPath(watchlistFilePath string) string {
	_, file := filepath.Split(watchlistFilePath)
	u.logger.Infof("Getting id from path '%s' - '%s'", watchlistFilePath, file)
	return file
}

func (u *Update) buildItemIdSlice(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return []string{}, err
	}

	return doc.Find("#DataTable tbody tr").Map(func(_ int, s *goquery.Selection) string {
		return s.AttrOr("id", "")
	}), nil
}

func (u *Update) getResultID(path string) (string, error) {
	ids, err := u.buildItemIdSlice(path)
	if err != nil || len(ids) == 0 {
		return "", err
	}

	sort.Strings(ids)
	return fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))), nil
}
