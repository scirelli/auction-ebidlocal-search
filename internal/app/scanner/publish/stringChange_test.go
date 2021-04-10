package publish

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStringChange(t *testing.T) {
	var l StringPublisher = NewStringChange()
	assert.NotNilf(t, l, "New should return a Listeners", l)
}

func TestStringChangeSubscribe(t *testing.T) {
	t.Run("Test subscribe one listener channel", func(t *testing.T) {
		var l StringPublisher = NewStringChange()
		var ch <-chan string
		var f func() error

		ch, f = l.Subscribe()
		assert.NotNilf(t, ch, "Calling subscribe once should subscribe the chan", ch)
		assert.NotNilf(t, f, "should return a cancel function", f)
	})
	t.Run("Test subscribe two listener channel", func(t *testing.T) {
		var l StringPublisher = NewStringChange()
		var ch <-chan string
		var f func() error

		ch, f = l.Subscribe()
		assert.NotNilf(t, ch, "Calling subscribe once should subscribe the chan", ch)
		assert.NotNilf(t, f, "should return a cancel function", f)

		ch, f = l.Subscribe()
		assert.NotNilf(t, ch, "Calling subscribe once should subscribe the chan", ch)
		assert.NotNilf(t, f, "should return a cancel function", f)
	})
}

func TestStringChangeUnsubscribe(t *testing.T) {
	t.Run("Test unsubscribing one listener channel", func(t *testing.T) {
		var l StringPublisher = NewStringChange()

		ch, f := l.Subscribe()
		assert.Nil(t, f())
		v, ok := <-ch
		assert.Zerof(t, v, "Channel should be empty", v)
		assert.Falsef(t, ok, "Channel should be closed", ok)
	})
}

func TestStringPublish(t *testing.T) {
	t.Run("Test channel is notified of a change", func(t *testing.T) {
		var l StringPublisher = NewStringChange()
		var ch1 <-chan string
		var wl = "hi"
		var v string

		ch1, _ = l.Subscribe()

		l.Publish(wl)

		v = <-ch1

		assert.Equalf(t, wl, v, "channel should received the published string")
	})

	t.Run("Test multiple channels are notified of a change", func(t *testing.T) {
		var l StringPublisher = NewStringChange()
		var wl = "hi"
		var v string

		ch1, _ := l.Subscribe()
		ch2, _ := l.Subscribe()

		l.Publish(wl)

		v = <-ch1
		assert.Equalf(t, wl, v, "channel should received the published string")
		v = <-ch2
		assert.Equalf(t, wl, v, "channel should received the published string")
	})
}
