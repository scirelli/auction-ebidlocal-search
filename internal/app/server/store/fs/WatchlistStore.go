package fs

import (
	"errors"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func NewWatchlistStore(ebidlocal *ebidlocal.Ebidlocal, logger *log.Logger) *EbidlocalAsWatchlistStore {
	return &EbidlocalAsWatchlistStore{
		ebidlocal: ebidlocal,
		logger:    logger,
	}
}

type EbidlocalAsWatchlistStore struct {
	ebidlocal *ebidlocal.Ebidlocal
	logger    *log.Logger
}

func (wl *EbidlocalAsWatchlistStore) SaveWatchlist(watchlist *model.Watchlist) error {
	if err := wl.ebidlocal.AddWatchlist(watchlist.List); err != nil {
		return err
	}
	wl.ebidlocal.EnqueueWatchlist(watchlist.List)
	return nil
}

func (wl *EbidlocalAsWatchlistStore) LoadWatchlist(watchlistID string) (*model.Watchlist, error) {
	return nil, errors.New("Not implemented")
}

func (wl *EbidlocalAsWatchlistStore) DeleteWatchlist(watchlistID string) error {
	return errors.New("Not implemented")
}
