package publish

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//New create a new string Publisher.
func NewStringChange() StringPublisher {
	var logger = log.New("Publisher")
	return &StringChange{
		logger: logger,
	}
}

//StringChange implements the Notifiers interface.
type StringChange struct {
	listeners []chan<- string
	mu        sync.RWMutex
	logger    *log.Logger
}

//Subscribe creates a channel to listen for string changes, returns that channel and a function to unsubscribe it.
func (l *StringChange) Subscribe() (readChan <-chan string, unsubscribe func() error) {
	var stringChan = make(chan string)
	readChan = stringChan

	l.mu.Lock()
	l.listeners = append(l.listeners, stringChan)
	l.mu.Unlock()

	return readChan, func() error {
		return l.unsubscribe(stringChan)
	}
}

func (l *StringChange) unsubscribe(stringChan chan<- string) error {
	defer l.mu.Unlock()

	l.mu.Lock()
	for i, c := range l.listeners {
		if c == stringChan {
			l.listeners[i] = l.listeners[len(l.listeners)-1]
			l.listeners = l.listeners[:len(l.listeners)-1]
			close(stringChan)
			return nil
		}
	}

	return errors.New("Not found")
}

//Publish publishes the new string. Sends to all listening channels, one goroutine each. To help prevent goroutine leaks there is a 50ms timeout context used.
func (l *StringChange) Publish(s string) {
	defer l.mu.RUnlock()
	l.mu.RLock()
	for _, c := range l.listeners {
		duration := 50 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		go func(ctx context.Context, c chan<- string, cancel context.CancelFunc) {
			select {
			case c <- s:
				cancel()
			case <-ctx.Done():
				l.logger.Error.Println("Publish timed out.", s)
			}
		}(ctx, c, cancel)
	}
}
