package notify

import "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"

func New() Notifier {
	var n Listeners = Listeners(make([]chan<- watchlist.Watchlist, 1))
	return &n
}

type Listeners []chan<- watchlist.Watchlist

func (l *Listeners) Register(watchlistChan chan<- watchlist.Watchlist) {
	*l = append(*l, watchlistChan)
}

func (l *Listeners) Unregister(watchlistChan chan<- watchlist.Watchlist) {
	var listeners = *l
	for i, c := range listeners {
		if c == watchlistChan {
			listeners[i] = listeners[len(listeners)-1]
			listeners = listeners[:len(listeners)-1]
			break
		}
	}
	*l = listeners
}

func (l *Listeners) Notify(wl watchlist.Watchlist) {
	for _, c := range *l {
		select {
		case c <- wl:
		default:
		}
	}
}
