package update

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	}).ParseFiles(config.TemplateFile)

	if err != nil {
		logger.Fatal(err)
	}

	return &Update{
		config:          config,
		logger:          logger,
		template:        t,
		auctionSearcher: search.AuctionSearchFunc(search.SearchAuctions),
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
	ctx             context.Context
	changePublsr    publish.StringPublisher
}

//SetOpenAuctions sets the list of open auctions to search.
func (u *Update) SetOpenAuctions(openAuctions stringiter.Iterable) *Update {
	u.openAuctions = openAuctions
	return u
}

//SubscribeForChange returns a channel that can be monitored for changes, it also returns a function to call unsubscribe the channel.
func (u *Update) SubscribeForChange() (<-chan string, func() error) {
	return u.changePublsr.Subscribe()
}

//Update starts the batch update of watch lists, reading from watchlistFilePaths channel and enqueuing them to be updated.
func (u *Update) Update(watchlistFilePaths <-chan string) error {
	for path := range watchlistFilePaths {
		select {
		case <-u.ctx.Done():
			break
		default:
		}
		if err := u.updateWatchlistContent(watchlistIDFromPath(filepath.Dir(path))); err != nil {
			u.logger.Error(err)
			continue
		}
	}
	return nil
}

//updateWatchlistContent determines if a watch list's content has changed, updates that content then publishes that there was a change.
func (u *Update) updateWatchlistContent(id string) error {
	var err error
	var newContent bytes.Buffer

	u.logger.Infof("Updating watch list id: '%s'", id)
	if err = u.searchAuctionForWatchlist(id, &newContent); err != nil {
		return err
	}
	contentID := u.getSavedContentId(id)
	newContentID := getContentId(bytes.NewReader(newContent.Bytes()))
	if contentID == newContentID {
		return nil
	}

	u.saveContent(id, bytes.NewReader(newContent.Bytes()))
	u.saveContentHash(id, newContentID)
	u.changePublsr.Publish(id)

	return nil
}

func (u *Update) searchAuctionForWatchlist(id string, out io.Writer) error {
	watchlist, err := u.store.LoadWatchlist(u.ctx, id)
	if err != nil {
		u.logger.Error(err)
		return err
	}

	if err := u.template.Execute(out, struct {
		Rows          chan string
		WatchlistLink string
		WatchlistName string
	}{
		Rows:          u.auctionSearcher.Search(stringiter.SliceStringIterator(watchlist), u.openAuctions),
		WatchlistLink: u.config.ServerUrl + "/watchlist/" + id,
		WatchlistName: "<!--{{watchlistName}}-->",
	}); err != nil {
		return err
	}

	return nil
}

func (u *Update) saveContent(watchlistID string, content io.Reader) error {
	var file *os.File
	var err error

	if file, err = os.Create(u.watchlistDataFilePathFromID(watchlistID)); err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, content)
	return err
}

func (u *Update) saveContentHash(watchlistID string, contentHash string) error {
	var file *os.File
	var err error

	if file, err = os.OpenFile(u.watchlistHashFilePathFromID(watchlistID), os.O_RDWR|os.O_CREATE, 0644); err != nil {
		return err
	}
	defer file.Close()
	if cnt, err := file.WriteAt([]byte(contentHash), 0); err != nil || cnt != len(contentHash) {
		return err
	}

	return nil
}

func (u *Update) getSavedContentId(watchlistID string) string {
	var file *os.File
	var err error
	if file, err = os.OpenFile(u.watchlistHashFilePathFromID(watchlistID), os.O_RDWR|os.O_CREATE, 0644); err != nil {
		u.logger.Error(err)
		return ""
	}
	defer file.Close()
	shaLength := 40
	buf := make([]byte, shaLength)
	if cnt, err := file.Read(buf); err != nil && err != io.EOF {
		u.logger.Errorf("Bytes read: '%d'; '%s'", cnt, err)
		return ""
	}
	return string(buf)
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

func watchlistIDFromPath(watchlistFilePath string) string {
	_, file := filepath.Split(watchlistFilePath)
	return file
}

func getContentId(content io.Reader) string {
	ids := buildItemIdSlice(content)
	if len(ids) == 0 {
		return ""
	}

	sort.Strings(ids)
	hash := fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, ""))))
	return hash
}

func buildItemIdSlice(content io.Reader) []string {
	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		return []string{}
	}

	return doc.Find("#DataTable tbody tr").Map(func(_ int, s *goquery.Selection) string {
		return s.AttrOr("id", "")
	})
}
