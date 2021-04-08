package publish

import "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"

type Publisher interface {
	Register() (<-chan watchlist.Watchlist, func() error)
	Publish(watchlist.Watchlist)
}
