package watchlist

import (
	"testing"
)

func TestWatchlistID(t *testing.T) {
	var w = Watchlist{"D", "R", "G", "Q", "A"}
	var expected = "HGKNlTs0ana3IDwNSaID7Sfve6U="

	if expected != w.ID() {
		t.Errorf("Failed ID creation got '%v'", w.ID())
	}
}
