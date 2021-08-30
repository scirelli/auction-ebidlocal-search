package ebidlocal

import "github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"

type AuctionSearcher interface {
	Search(keywords stringiter.Iterable, auctions stringiter.Iterable) (results chan string)
}
