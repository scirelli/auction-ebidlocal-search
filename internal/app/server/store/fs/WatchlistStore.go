package fs

import (
	"context"
	"errors"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	wlmodel "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	ebidstore "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//NewWatchlistStore adapter for Ebidlocal store.
func NewWatchlistStore(store ebidstore.Storer, logger log.Logger) *EbidlocalAsWatchlistStore {
	return &EbidlocalAsWatchlistStore{
		store:  store,
		logger: logger,
	}
}

//EbidlocalAsWatchlistStore adapter for Ebidlocal store.
type EbidlocalAsWatchlistStore struct {
	store  ebidstore.Storer
	logger log.Logger
}

func (wl *EbidlocalAsWatchlistStore) SaveWatchlist(ctx context.Context, list *model.Watchlist) (ID string, err error) {
	if ID, err = wl.store.SaveWatchlist(ctx, wlmodel.Watchlist(list.List)); err != nil {
		return "", err
	}
	return ID, nil
}

func (wl *EbidlocalAsWatchlistStore) LoadWatchlist(ctx context.Context, watchlistID string) (*model.Watchlist, error) {
	return nil, errors.New("Not implemented")
}

func (wl *EbidlocalAsWatchlistStore) DeleteWatchlist(ctx context.Context, watchlistID string) error {
	return errors.New("Not implemented")
}
