package store

import (
	"context"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"
)

//Storer interface to represent a data store.
type Storer interface {
	WatchlistStorer
}

//WatchlistStorer storeable to perform watchlist store operations.
type WatchlistStorer interface {
	SaveWatchlist(ctx context.Context, watchlist watchlist.Watchlist) (string, error)
	LoadWatchlist(ctx context.Context, watchlistID string) (watchlist.Watchlist, error)
	DeleteWatchlist(ctx context.Context, watchlistID string) error
}
