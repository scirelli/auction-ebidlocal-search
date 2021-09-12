package model

import (
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
