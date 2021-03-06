package publish

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSliceStringChange(t *testing.T) {
	var l SliceStringPublisher = NewSliceStringChange()
	assert.NotNilf(t, l, "New should return a Listeners", l)
}

func TestSliceStringChangeRegister(t *testing.T) {
	t.Run("Test registering one listener channel", func(t *testing.T) {
		var l SliceStringPublisher = NewSliceStringChange()
		var ch <-chan []string
		var f func() error

		ch, f = l.Register()
		assert.NotNilf(t, ch, "Calling register once should register the chan", ch)
		assert.NotNilf(t, f, "should return a cancel function", f)
	})
	t.Run("Test registering two listener channel", func(t *testing.T) {
		var l SliceStringPublisher = NewSliceStringChange()
		var ch <-chan []string
		var f func() error

		ch, f = l.Register()
		assert.NotNilf(t, ch, "Calling register once should register the chan", ch)
		assert.NotNilf(t, f, "should return a cancel function", f)

		ch, f = l.Register()
		assert.NotNilf(t, ch, "Calling register once should register the chan", ch)
		assert.NotNilf(t, f, "should return a cancel function", f)
	})
}

func TestSliceStringChangeUnregister(t *testing.T) {
	t.Run("Test unregistering one listener channel", func(t *testing.T) {
		var l SliceStringPublisher = NewSliceStringChange()

		ch, f := l.Register()
		assert.Nil(t, f())
		v, ok := <-ch
		assert.Zerof(t, v, "Channel should be empty", v)
		assert.Falsef(t, ok, "Channel should be closed", ok)
	})
}

func TestSliceStringPublish(t *testing.T) {
	t.Run("Test channel is notified of a change", func(t *testing.T) {
		var l SliceStringPublisher = NewSliceStringChange()
		var ch1 <-chan []string
		var wl = []string([]string{"a", "b"})
		var v []string

		ch1, _ = l.Register()

		l.Publish(wl)

		v = <-ch1

		assert.Equalf(t, wl, v, "channel should received the published watch list")
	})

	t.Run("Test multiple channels are notified of a change", func(t *testing.T) {
		var l SliceStringPublisher = NewSliceStringChange()
		var wl = []string([]string{"a", "b"})
		var v []string

		ch1, _ := l.Register()
		ch2, _ := l.Register()

		l.Publish(wl)

		v = <-ch1
		assert.Equalf(t, wl, v, "channel should received the published watch list")
		v = <-ch2
		assert.Equalf(t, wl, v, "channel should received the published watch list")
	})
}
