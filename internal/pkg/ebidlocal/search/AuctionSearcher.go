package search

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
)

type AuctionSearcher interface {
	Search(keywords stringiter.Iterable) (results chan model.SearchResult)
}
