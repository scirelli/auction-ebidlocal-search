package notify

import "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"

type Notifier interface {
	Register(chan<- watchlist.Watchlist)
	Unregister(chan<- watchlist.Watchlist)
	Notify(watchlist.Watchlist)
}
