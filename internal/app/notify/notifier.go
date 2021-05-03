package notify

import (
	"fmt"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
)

type NotificationMessage struct {
	User        *model.User
	WatchlistID string
}

func (n NotificationMessage) String() string {
	return fmt.Sprintf("User: %s; Watch list: %s", n.User, n.WatchlistID)
}

type WatchlistNotifier interface {
	Notify(message NotificationMessage) error
}

type NotifyFunc func(message NotificationMessage) error

func (nf NotifyFunc) Notify(message NotificationMessage) error {
	return nf(message)
}

func NotifyAll(messages <-chan NotificationMessage, notifiers ...WatchlistNotifier) {
	for msg := range messages {
		for _, n := range notifiers {
			go func(n WatchlistNotifier) {
				n.Notify(msg)
			}(n)
		}
	}
}
