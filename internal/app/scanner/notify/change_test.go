package notify

import (
	"testing"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/watchlist"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var l Notifier = New()
	assert.NotNilf(t, l, "New should return a Listeners", l)
}

func TestRegister(t *testing.T) {
	var l Listeners = Listeners{}

	t.Run("Test registering one listener channel", func(t *testing.T) {
		var ch = make(chan watchlist.Watchlist)
		var sendOnly chan<- watchlist.Watchlist = ch
		l.Register(ch)
		assert.Equalf(t, sendOnly, l[0], "Calling register once should register the chan", ch)
	})
	t.Run("Test registering another listener channel", func(t *testing.T) {
		var ch = make(chan watchlist.Watchlist)
		var sendOnly chan<- watchlist.Watchlist = ch
		l.Register(ch)
		assert.Equalf(t, sendOnly, l[1], "Calling register once should register the chan", ch)
	})
}

func TestUnregister(t *testing.T) {
	var l Listeners = Listeners{}

	t.Run("Test unregister one listener channel", func(t *testing.T) {
		var ch1 = make(chan watchlist.Watchlist)
		l.Register(ch1)
		assert.Equalf(t, len(l), 1, "Should register 1 channels", len(l))
		l.Unregister(ch1)
		assert.Equalf(t, len(l), 0, "calling unregister should remove the provided listener", len(l))
	})
	t.Run("Test unregister another listener channel", func(t *testing.T) {
		var ch1 = make(chan watchlist.Watchlist)
		var ch2 = make(chan watchlist.Watchlist)
		l.Register(ch1)
		assert.Equalf(t, len(l), 1, "Should register 1 channels", len(l))
		l.Register(ch2)
		assert.Equalf(t, len(l), 2, "Should register 2 channels", len(l))
		l.Unregister(ch2)
		assert.Equalf(t, len(l), 1, "calling unregister should remove the provided listener", len(l))
		l.Unregister(ch1)
		assert.Equalf(t, len(l), 0, "calling unregister should remove the provided listener", len(l))
	})

	t.Run("Test unregister a listener that was not registered", func(t *testing.T) {
		var ch1 = make(chan watchlist.Watchlist)
		var ch2 = make(chan watchlist.Watchlist)

		l.Unregister(ch2)
		assert.Equalf(t, len(l), 0, "calling unregister on a channel that was not registered should do nothing.", len(l))

		l.Register(ch2)
		assert.Equalf(t, len(l), 1, "Should register 1 channels", len(l))
		l.Unregister(ch1)
		assert.Equalf(t, len(l), 1, "should do nothing if unregistering a channel that was not registered.", len(l))
	})
}

func TestNotify(t *testing.T) {
	t.Run("Test channel is notified of a change", func(t *testing.T) {
		var l Listeners = Listeners{}
		var ch1 = make(chan watchlist.Watchlist, 1)
		var wl = watchlist.Watchlist([]string{"a", "b"})
		var v watchlist.Watchlist

		l.Register(ch1)

		l.Notify(wl)

		select {
		case v = <-ch1:
		default:
		}

		assert.Equalf(t, v, wl, "channel should received the notified watch list")
	})

	t.Run("Test closed channel is not notified", func(t *testing.T) {
		var l Listeners = Listeners{}
		var ch1 = make(chan watchlist.Watchlist)
		var wl = watchlist.Watchlist([]string{"a", "b"})
		var v watchlist.Watchlist

		l.Register(ch1)
		close(ch1)

		l.Notify(wl)

		select {
		case v = <-ch1:
		default:
		}

		assert.Nilf(t, v, "channel should received the notified watch list")
	})
}
