package watchlist

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/id"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
)

type Watchlister interface {
	id.IDer
	stringiter.Iterable
}
