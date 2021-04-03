package notify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var l Notifier = New()
	assert.NotNilf(t, l, "New should return a Listeners", l)
}
