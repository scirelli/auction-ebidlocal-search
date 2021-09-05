package ebidlocal

import "github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"

type AuctionSearchFunc func(keywords stringiter.Iterable) (results chan string)

func (as AuctionSearchFunc) Search(keywords stringiter.Iterable) (results chan string) {
	return as(keywords)
}
