package ebidlocal

import (
	"crypto/sha1"
	"sort"
	"strings"
)

type Watchlist []string

func (w Watchlist) ID() string {
	sort.Sort(w)
	h := sha1.New()
	for _, s := range w {
		h.Write([]byte(strings.ToLower(s)))
	}
	return strings.ToUpper(string(h.Sum(nil)))
}

// Len is the number of elements in the collection.
func (w Watchlist) Len() int {
	return len(w)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (w Watchlist) Less(i, j int) bool {
	return strings.ToLower(w[i]) < strings.ToLower(w[j])
}

// Swap swaps the elements with indexes i and j.
func (w Watchlist) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}
