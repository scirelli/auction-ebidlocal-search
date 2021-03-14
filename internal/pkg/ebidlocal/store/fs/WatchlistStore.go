package fs

import (
	"context"
	"errors"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal/watchlist"
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

func (wl *EbidlocalAsWatchlistStore) SaveWatchlist(ctx context.Context, wlist *model.Watchlist) (ID string, err error) {
	if err = wl.ebidlocal.AddWatchlist(wlist.List); err != nil {
		return "", err
	}
	wl.ebidlocal.EnqueueWatchlist(wlist.List)
	return watchlist.Watchlist(wlist.List).ID(), nil
}

func (wl *EbidlocalAsWatchlistStore) LoadWatchlist(ctx context.Context, watchlistID string) (*model.Watchlist, error) {
	return nil, errors.New("Not implemented")
}

func (wl *EbidlocalAsWatchlistStore) DeleteWatchlist(ctx context.Context, watchlistID string) error {
	return errors.New("Not implemented")
}
