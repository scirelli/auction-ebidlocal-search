package publish

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//NewSliceStringChange create a new SliceStringChange Publisher.
func NewSliceStringChange() SliceStringPublisher {
	var logger = log.New("Publisher")
	return &SliceStringChange{
		logger: logger,
	}
}

//SliceStringChange implements the Notifiers interface.
type SliceStringChange struct {
	listeners []chan<- []string
	mu        sync.RWMutex
	logger    *log.Logger
}

//Register creates a channel to listen for slice string changes, returns that channel and a function to unregister it.
func (l *SliceStringChange) Register() (readChan <-chan []string, unregister func() error) {
	var sliceStringChan = make(chan []string)
	readChan = sliceStringChan

	l.mu.Lock()
	l.listeners = append(l.listeners, sliceStringChan)
	l.mu.Unlock()

	return readChan, func() error {
		return l.unregister(sliceStringChan)
	}
}

func (l *SliceStringChange) unregister(sliceStringChan chan<- []string) error {
	defer l.mu.Unlock()

	l.mu.Lock()
	for i, c := range l.listeners {
		if c == sliceStringChan {
			l.listeners[i] = l.listeners[len(l.listeners)-1]
			l.listeners = l.listeners[:len(l.listeners)-1]
			close(sliceStringChan)
			return nil
		}
	}

	return errors.New("Not found")
}

//Publish publishes the new watch list. Sends to all listening channels, one goroutine each. To help prevent goroutine leaks there is a 50ms timeout context used.
func (l *SliceStringChange) Publish(wl []string) {
	defer l.mu.RUnlock()
	l.mu.RLock()
	for _, c := range l.listeners {
		duration := 50 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		go func(ctx context.Context, c chan<- []string, cancel context.CancelFunc) {
			select {
			case c <- wl:
				cancel()
			case <-ctx.Done():
				l.logger.Error.Println("Publish timed out.", wl)
			}
		}(ctx, c, cancel)
	}
}
