package id

import (
	"crypto/sha1"
	b64 "encoding/base64"
	"sort"
	"strings"
)

type IDedSliceString []string

func (w IDedSliceString) ID() string {
	w.Normalize()
	h := sha1.New()
	for _, s := range w {
		h.Write([]byte(s))
	}
	return b64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (w IDedSliceString) Normalize() IDedSliceString {
	for i, s := range w {
		w[i] = strings.ToLower(s)
	}
	sort.Sort(w)

	return w
}

// Len is the number of elements in the collection.
func (w IDedSliceString) Len() int {
	return len(w)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (w IDedSliceString) Less(i, j int) bool {
	return strings.ToLower(w[i]) < strings.ToLower(w[j])
}

// Swap swaps the elements with indexes i and j.
func (w IDedSliceString) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}
