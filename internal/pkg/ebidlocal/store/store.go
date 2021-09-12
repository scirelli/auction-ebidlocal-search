package store

import (
	"context"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
)

//Storer interface to represent a data store.
type Storer interface {
	WatchlistStorer
	WatchlistContentStorer
}

//WatchlistStorer storeable to perform watchlist store operations.
type WatchlistStorer interface {
	SaveWatchlist(ctx context.Context, watchlist model.Watchlist) (string, error)
	LoadWatchlist(ctx context.Context, watchlistID string) (model.Watchlist, error)
	DeleteWatchlist(ctx context.Context, watchlistID string) error
}

//WatchlistContentStorer storeable to perform watchlist content store operations.
type WatchlistContentStorer interface {
	SaveWatchlistContent(ctx context.Context, watchlistContent *model.WatchlistContent) (string, error)
	LoadWatchlistContent(ctx context.Context, watchlistContentID string) (*model.WatchlistContent, error)
	DeleteWatchlistContent(ctx context.Context, watchlistContentID string) error
}
