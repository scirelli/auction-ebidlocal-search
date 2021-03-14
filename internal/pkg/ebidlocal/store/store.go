package store

import (
	"context"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
)

//Storer interface to represent a data store.
type Storer interface {
	UserStorer
	WatchlistStorer
}

//UserStorer store able to perform user store operations.
type UserStorer interface {
	SaveUser(ctx context.Context, u *model.User) (string, error)
	LoadUser(ctx context.Context, userID string) (*model.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

//WatchlistStorer store able to perform watchlist store operations.
type WatchlistStorer interface {
	SaveWatchlist(ctx context.Context, watchlist *model.Watchlist) (string, error)
	LoadWatchlist(ctx context.Context, watchlistID string) (*model.Watchlist, error)
	DeleteWatchlist(ctx context.Context, watchlistID string) error
}
