package ebidlocal

import (
	"crypto/sha1"
	"sort"
	"strings"
)

type Watchlist []string

func (w Watchlist) ID() string {
	w.Normalize()
	h := sha1.New()
	for _, s := range w {
		h.Write([]byte(s))
	}
	return strings.ToUpper(string(h.Sum(nil)))
}

func (w Watchlist) Normalize() Watchlist {
	for i, s := range w {
		w[i] = strings.ToLower(s)
	}
	sort.Sort(w)

	return w
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
