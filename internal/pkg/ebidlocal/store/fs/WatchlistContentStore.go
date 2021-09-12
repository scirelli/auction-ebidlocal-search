package fs

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func NewWatchlistContentStore(config WatchlistContentStoreConfig) *WatchlistContentStore {
	var logger = log.New("WatchlistContentStore", log.DEFAULT_LOG_LEVEL)

	if config.DataFileName == "" {
		config.DataFileName = "models.json"
	}
	if config.ContentPath == "" {
		config.ContentPath = "."
		logger.Infof("Defaulting content path dir to '%s'\n", config.ContentPath)
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Infof("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}

	return &WatchlistContentStore{
		Config: config,
		logger: logger,
	}
}

type WatchlistContentStore struct {
	Config WatchlistContentStoreConfig
	logger log.Logger
}

type WatchlistContentStoreConfig struct {
	ContentPath  string `json:"contentPath"`
	WatchlistDir string `json:"watchlistDir"`
	DataFileName string `json:"dataFileName"`
}

func (wc *WatchlistContentStore) SaveWatchlistContent(ctx context.Context, watchlistContent *model.WatchlistContent) (string, error) {
	var file *os.File
	var err error

	if file, err = os.Create(wc.watchlistDataFilePathFromID(watchlistContent.GetWatchlistID())); err != nil {
		return "", err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	if err = enc.Encode(watchlistContent); err != nil {
		return "", err
	}

	return watchlistContent.ID(), nil
}

func (wc *WatchlistContentStore) LoadWatchlistContent(ctx context.Context, watchlistContentID string) (*model.WatchlistContent, error) {
	var wcModel model.WatchlistContent
	var byteValue []byte
	var err error

	if byteValue, err = ioutil.ReadFile(wc.watchlistDataFilePathFromID(watchlistContentID)); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(byteValue, &wcModel); err != nil {
		return nil, err
	}

	return &wcModel, nil
	/* Streamed in
	var wcModel model.WatchlistContent
	var file *os.File
	var err error

	if file, err = os.Open(wc.watchlistDataFilePathFromID(watchlistContentID)); err != nil {
		return &model.WatchlistContent{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if _, err := decoder.Token(); err != nil {
		return &model.WatchlistContent{}, err
	}
	for decoder.More() {
		if err := decoder.Decode(&wcModel); err != nil {
			return &model.WatchlistContent{}, err
		}
	}

	return &wcModel, nil
	*/
}

func (wc *WatchlistContentStore) DeleteWatchlistContent(ctx context.Context, watchlistContentID string) error {
	return os.Remove(wc.watchlistDataFilePathFromID(watchlistContentID))
}

func (wc *WatchlistContentStore) watchlistDataFilePathFromID(watchlistID string) string {
	return filepath.Join(wc.Config.WatchlistDir, watchlistID, wc.Config.DataFileName)
}
