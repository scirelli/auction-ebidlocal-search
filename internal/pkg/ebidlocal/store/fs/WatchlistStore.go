package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func NewWatchlistStore(config StoreConfig, logger log.Logger) *WatchlistStore {
	return &WatchlistStore{
		config: config,
		logger: logger,
	}
}

type WatchlistStore struct {
	config StoreConfig
	logger log.Logger
}

type StoreConfig struct {
	WatchlistDir string `json:"watchlistDir"`
	DataFileName string `json:"dataFileName"`
}

func (wl *WatchlistStore) SaveWatchlist(ctx context.Context, list watchlist.Watchlist) (ID string, err error) {
	if err = wl.addWatchlist(list); err != nil {
		return "", err
	}
	return list.ID(), nil
}

func (wl *WatchlistStore) LoadWatchlist(ctx context.Context, watchlistID string) (watchlist.Watchlist, error) {
	return wl.loadWatchlist(filepath.Join(wl.config.WatchlistDir, watchlistID, wl.config.DataFileName))
}

func (wl *WatchlistStore) DeleteWatchlist(ctx context.Context, watchlistID string) error {
	return errors.New("Not implemented")
}

//AddWatchlist saves a watchlist to disk. Skips saving if it already exists.
func (wl *WatchlistStore) addWatchlist(list watchlist.Watchlist) error {
	var watchlistDir = filepath.Join(wl.config.WatchlistDir, list.ID())

	wl.logger.Infof("Checking for '%s'\n", watchlistDir)
	if _, err := os.Stat(watchlistDir); os.IsExist(err) {
		wl.logger.Info("Watch list already exists.")
		return nil
	}

	wl.logger.Infof("Creating watchlist. '%s'", watchlistDir)
	if err := os.MkdirAll(watchlistDir, 0775); err != nil {
		wl.logger.Error(err)
		return err
	}

	file, err := json.Marshal(list)
	if err != nil {
		wl.logger.Error(err)
		if err2 := os.RemoveAll(watchlistDir); err2 != nil {
			err = fmt.Errorf("%v: %w", err, err2)
		}
		return err
	}

	return ioutil.WriteFile(filepath.Join(watchlistDir, "data.json"), file, 0644)
}

//loadWatchlist loads a watch list from file.
func (wl *WatchlistStore) loadWatchlist(filePath string) (watchlist.Watchlist, error) {
	var watchlist watchlist.Watchlist = make([]string, 0)
	jsonFile, err := os.Open(filePath)
	if err != nil {
		wl.logger.Error(err)
		return watchlist, err
	}
	defer jsonFile.Close()
	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&watchlist); err != nil {
		wl.logger.Error(err)
		return watchlist, err
	}
	wl.logger.Info("Watch list found '%v'", watchlist)
	return watchlist, nil
}
