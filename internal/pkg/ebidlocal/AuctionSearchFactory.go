package ebidlocal

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	search "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search"
	v1 "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search/v1"
	v2 "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search/v2"
	v3 "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search/v3"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
)

type SearcherFactoryFunc func(interface{}) search.AuctionSearcher

var searchers map[string]SearcherFactoryFunc = map[string]SearcherFactoryFunc{
	"nil": func(config interface{}) search.AuctionSearcher {
		return search.AuctionSearchFunc(NullSearch)
	},
	"null": func(config interface{}) search.AuctionSearcher {
		return search.AuctionSearchFunc(NullSearch)
	},
	"NullSearch": func(config interface{}) search.AuctionSearcher {
		return search.AuctionSearchFunc(NullSearch)
	},
	"": func(config interface{}) search.AuctionSearcher {
		return search.AuctionSearchFunc(NullSearch)
	},
	"v3": func(config interface{}) search.AuctionSearcher {
		return search.AuctionSearchFunc(func(keywordIter stringiter.Iterable) chan model.SearchResult {
			return v3.SearchAuctions(keywordIter, v2.NewAuctionsCache())
		})
	},
	"v2": func(config interface{}) search.AuctionSearcher {
		return search.AuctionSearchFunc(func(keywordIter stringiter.Iterable) chan model.SearchResult {
			return v2.SearchAuctions(keywordIter, v2.NewAuctionsCache())
		})
	},
	"v1": func(config interface{}) search.AuctionSearcher {
		return search.AuctionSearchFunc(func(keywordIter stringiter.Iterable) chan model.SearchResult {
			return v1.SearchAuctions(keywordIter, v1.NewAuctionsCache())
		})
	},
}

func AuctionSearchFactory(version string, config interface{}) search.AuctionSearcher {
	if searcher, ok := searchers[version]; ok {
		return searcher(config)
	}
	return search.AuctionSearchFunc(NullSearch)
}

func AuctionSearchRegistrar(name string, f SearcherFactoryFunc) {
	searchers[name] = f
}

func NullSearch(keywords stringiter.Iterable) chan model.SearchResult {
	c := make(chan model.SearchResult)
	close(c)
	return c
}
