package search

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
)

type AuctionSearchFunc func(keywords stringiter.Iterable) (results chan model.SearchResult)

func (as AuctionSearchFunc) Search(keywords stringiter.Iterable) (results chan model.SearchResult) {
	return as(keywords)
}
