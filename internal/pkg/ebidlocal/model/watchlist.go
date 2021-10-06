package model

import (
	"sort"
	"strings"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/id"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
)

type Watchlist []string

func (w Watchlist) ID() string {
	return id.IDedSliceString(w).ID()
}

func (w Watchlist) Iterator() stringiter.Iterator {
	return stringiter.SliceStringIterator(w).Iterator()
}

func (w *Watchlist) Normalize() *Watchlist {
	set := make(map[string]struct{})
	for _, s := range *w {
		set[s] = struct{}{}
	}

	i := 0
	tmp := Watchlist(make([]string, len(set)))
	for key := range set {
		tmp[i] = strings.ToLower(key)
		i++
	}

	*w = tmp
	sort.Strings(*w)
	return w
}
