package ebidlocal

import (
	"strings"
	"testing"
)

func TestWatchlistID(t *testing.T) {
	var w = Watchlist{"D", "R", "G", "Q", "A"}
	var expected = strings.ToUpper(string([]byte{28, 98, 141, 149, 59, 52, 106, 118, 183, 32, 60, 13, 73, 162, 3, 237, 39, 239, 123, 165}))

	if expected != w.ID() {
		t.Errorf("Failed ID creation")
	}
}
