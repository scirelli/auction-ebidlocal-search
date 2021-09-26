package update

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/filter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/publish"
)

type Updater interface {
	Update(watchlistPath <-chan string) error
}

//New constructor for updater app. The updater subscribes to watch list file channel. When it receives a watch list it then updates the data.
func New(ctx context.Context, watchlistStore store.Storer, searchExtractor SearchExtractor, config Config) *Update {
	var logger = log.New("Update", log.DEFAULT_LOG_LEVEL)

	return &Update{
		config:          config,
		logger:          logger,
		searchExtractor: searchExtractor,
		ctx:             ctx,
		store:           watchlistStore,
		changePublsr:    publish.NewStringChange(),
	}
}

//Update data
type Update struct {
	config          Config
	logger          log.Logger
	searchExtractor SearchExtractor
	store           store.Storer
	ctx             context.Context
	changePublsr    publish.StringPublisher
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
			u.logger.Debug("Update.Update: ctx done, ending update checks.")
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
	var watchlist model.Watchlist
	var watchlistContent = model.WatchlistContent{
		WatchlistID: id,
		Timestamp:   time.Now(),
	}

	if watchlist, err = u.store.LoadWatchlist(u.ctx, id); err != nil {
		u.logger.Error(err)
		return err
	}

	u.logger.Debugf("Updater.updateWatchlistContent: Checking watch list id: '%s'", id)
	for item := range u.searchAuctionForWatchlist(watchlist) {
		watchlistContent.AuctionItems = append(watchlistContent.AuctionItems, item)
	}

	contentID := u.getSavedContentId(id)
	if contentID == watchlistContent.ID() {
		u.logger.Debugf("Updater.updateWatchlistContent: No changes for id('%s')", contentID)
		return nil
	}

	u.logger.Debugf("Updater.updateWatchlistContent: There was a change to watch list: '%s'", id)
	if _, err = u.store.SaveWatchlistContent(u.ctx, &watchlistContent); err != nil {
		u.logger.Debugf("Updater.saveContent: Was not able to save the content for watchlist '%s'", id)
		return err
	}
	if err = u.saveContentHash(id, watchlistContent.ID()); err != nil {
		u.logger.Debugf("Updater.saveContentHash: Was not able to save the content hash for watchlist '%s'", id)
		return err
	}
	u.logger.Debugf("Updater.updateWatchlistContent: Publishing change for: '%s'", id)
	u.changePublsr.Publish(id)

	return nil
}

func (u *Update) searchAuctionForWatchlist(watchlist model.Watchlist) <-chan model.AuctionItem {
	return model.FilterAuctionItemChan(u.searchExtractor.Extract(u.searchExtractor.Search(stringiter.SliceStringIterator(watchlist)))).Filter(model.FilterFunc(filter.ByKeyword))
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

func (u *Update) watchlistHashFilePathFromID(watchlistID string) string {
	return filepath.Join(u.config.WatchlistDir, watchlistID, "hash")
}

func watchlistIDFromPath(watchlistFilePath string) string {
	_, file := filepath.Split(watchlistFilePath)
	return file
}

/* Stream models out to file

enc := json.NewEncoder(out)
if _, err := w.Write([]byte{'['}); err != nil {
	u.logger.Error(err)
	return err
}
if err := enc.Encode(<-models); err != nil {
	if _, err := w.Write([]byte{']'}); err != nil {
		u.logger.Error(err)
		return err
	}
	return err
}
for o := range models {
	if _, err := w.Write([]byte{','}); err != nil {
		return err
	}
	if err := enc.Encode(o); err != nil {
		if _, err := w.Write([]byte{']'}); err != nil {
			u.logger.Error(err)
			return err
		}
		return err
	}
}
return nil
*/
