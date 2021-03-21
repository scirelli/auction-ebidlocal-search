package notify

func New() Notifier {
	return &Change{}
}

type Change struct {
	listeners []chan Watchlist
}

func (c *Change) Register(watchlistChan chan<- Watchlist) {
	c.listeners = append(c.listeners, watchlistChan)
}

func (c *Change) Unregister(watchlistChan chan<- Watchlist) {
	for i, c := range c.listeners {
		if c == watchlistChan {
			e.listeners[i] = e.listeners[len(e.listeners)-1]
			e.listeners = e.listeners[:len(e.listeners)-1]
			break
		}
	}
}

func (c *Change) Notify(wl Watchlist) {
	for _, c := range c.listeners {
		select {
		case c <- wl:
		default:
		}
	}
}
