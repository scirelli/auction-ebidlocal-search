package publish

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//New create a new watchlist Publisher.
func New() Publisher {
	var logger = log.New("Publisher")
	return &Listeners{
		logger: logger,
	}
}

//Listeners implements the Notifiers interface.
type Listeners struct {
	listeners []chan<- watchlist.Watchlist
	mu        sync.Mutex
	logger    *log.Logger
}

//Register creates a channel to listen for watchlist changes, returns that channel and a function to unregister it.
func (l *Listeners) Register() (readChan <-chan watchlist.Watchlist, unregister func() error) {
	var watchlistChan = make(chan watchlist.Watchlist)
	readChan = watchlistChan

	l.mu.Lock()
	l.listeners = append(l.listeners, watchlistChan)
	l.mu.Unlock()

	return readChan, func() error {
		return l.unregister(watchlistChan)
	}
}

func (l *Listeners) unregister(watchlistChan chan<- watchlist.Watchlist) error {
	defer l.mu.Unlock()

	l.mu.Lock()
	for i, c := range l.listeners {
		if c == watchlistChan {
			l.listeners[i] = l.listeners[len(l.listeners)-1]
			l.listeners = l.listeners[:len(l.listeners)-1]
			close(watchlistChan)
			return nil
		}
	}

	return errors.New("Not found")
}

//Publish publishes the new watch list. Sends to all listening channels, one goroutine each. To help prevent goroutine leaks there is a 50ms timeout context used.
func (l *Listeners) Publish(wl watchlist.Watchlist) {
	defer l.mu.Unlock()
	l.mu.Lock()
	for _, c := range l.listeners {
		duration := 50 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		go func(ctx context.Context, c chan<- watchlist.Watchlist, cancel context.CancelFunc) {
			select {
			case c <- wl:
				cancel()
			case <-ctx.Done():
				l.logger.Error.Println("Publish timed out.", wl)
			}
		}(ctx, c, cancel)
	}
}
