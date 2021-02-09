package store

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
)

//Storer interface to represent a data store.
type Storer interface {
	UserStorer
	WatchlistStorer
}

//UserStorer store able to perform user store operations.
type UserStorer interface {
	SaveUser(u *model.User) error
	LoadUser(userID string) (*model.User, error)
	DeleteUser(userID string) error
}

//WatchlistStorer store able to perform watchlist store operations.
type WatchlistStorer interface {
	SaveWatchlist(watchlist *model.Watchlist) error
	LoadWatchlist(watchlistID string) (*model.Watchlist, error)
	DeleteWatchlist(watchlistID string) error
}
