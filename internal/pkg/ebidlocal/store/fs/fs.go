package fs

import "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store"

type FSStore struct {
	store.WatchlistStorer
	store.WatchlistContentStorer
}
