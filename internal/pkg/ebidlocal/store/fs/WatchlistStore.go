package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func NewWatchlistStore(config WatchlistStoreConfig, logger log.Logger) *WatchlistStore {
	if logger == nil {
		logger = log.New("WatchlistStore", log.DEFAULT_LOG_LEVEL)
	}

	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}

	return &WatchlistStore{
		Config: config,
		Logger: logger,
	}
}

type WatchlistStore struct {
	Config WatchlistStoreConfig
	Logger log.Logger
}

type WatchlistStoreConfig struct {
	WatchlistDir string `json:"watchlistDir"`
	DataFileName string `json:"dataFileName"`
}

func (wl *WatchlistStore) SaveWatchlist(ctx context.Context, list model.Watchlist) (ID string, err error) {
	if err = wl.addWatchlist(list); err != nil {
		return "", err
	}
	return list.ID(), nil
}

func (wl *WatchlistStore) LoadWatchlist(ctx context.Context, watchlistID string) (model.Watchlist, error) {
	return wl.loadWatchlist(filepath.Join(wl.Config.WatchlistDir, watchlistID, wl.Config.DataFileName))
}

func (wl *WatchlistStore) DeleteWatchlist(ctx context.Context, watchlistID string) error {
	return errors.New("Not implemented")
}

//AddWatchlist saves a watchlist to disk. Skips saving if it already exists.
func (wl *WatchlistStore) addWatchlist(list model.Watchlist) error {
	var watchlistDir = filepath.Join(wl.Config.WatchlistDir, list.ID())

	wl.Logger.Infof("WatchlistStore.addWatchlist: Checking for '%s'\n", watchlistDir)
	if _, err := os.Stat(watchlistDir); os.IsExist(err) {
		wl.Logger.Info("WatchlistStore.addWatchlist: Watch list already exists.")
		return nil
	}

	wl.Logger.Infof("WatchlistStore.addWatchlist: Creating watchlist. '%s'", watchlistDir)
	if err := os.MkdirAll(watchlistDir, 0775); err != nil {
		wl.Logger.Error(err)
		return err
	}

	file, err := json.Marshal(list)
	if err != nil {
		wl.Logger.Error(err)
		if err2 := os.RemoveAll(watchlistDir); err2 != nil {
			err = fmt.Errorf("%v: %w", err, err2)
		}
		return err
	}

	return ioutil.WriteFile(filepath.Join(watchlistDir, wl.Config.DataFileName), file, 0644)
}

//loadWatchlist loads a watch list from file.
func (wl *WatchlistStore) loadWatchlist(filePath string) (model.Watchlist, error) {
	var watchlist model.Watchlist = make([]string, 0)
	jsonFile, err := os.Open(filePath)
	if err != nil {
		wl.Logger.Error(err)
		return watchlist, err
	}
	defer jsonFile.Close()
	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&watchlist); err != nil {
		wl.Logger.Error(err)
		return watchlist, err
	}
	wl.Logger.Infof("WatchlistStore.loadWatchlist: Watch list found '%v'", watchlist)
	return watchlist, nil
}
