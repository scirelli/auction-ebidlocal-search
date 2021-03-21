package notify

type Notifier interface {
	Register(chan<- Watchlist)
	Unregister(chan<- Watchlist)
	Notify(Watchlist)
}
