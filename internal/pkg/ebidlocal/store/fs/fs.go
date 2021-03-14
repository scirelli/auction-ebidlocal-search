package fs

import "github.com/scirelli/auction-ebidlocal-search/internal/app/server/store"

type FSStore struct {
	store.UserStorer
	store.WatchlistStorer
}
