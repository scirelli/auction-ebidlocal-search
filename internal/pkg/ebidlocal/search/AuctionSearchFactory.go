package search

import "github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"

type SearcherFactoryFunc func(interface{}) AuctionSearcher

var searchers map[string]SearcherFactoryFunc = map[string]SearcherFactoryFunc{
	"nil": func(config interface{}) AuctionSearcher {
		return AuctionSearchFunc(NullSearch)
	},
	"null": func(config interface{}) AuctionSearcher {
		return AuctionSearchFunc(NullSearch)
	},
	"NullSearch": func(config interface{}) AuctionSearcher {
		return AuctionSearchFunc(NullSearch)
	},
	"": func(config interface{}) AuctionSearcher {
		return AuctionSearchFunc(NullSearch)
	},
}

func AuctionSearchFactory(version string, config interface{}) AuctionSearcher {
	if searcher, ok := searchers[version]; ok {
		return searcher(config)
	}
	return AuctionSearchFunc(NullSearch)
}

func AuctionSearchRegistrar(name string, f SearcherFactoryFunc) {
	searchers[name] = f
}

func NullSearch(keywords stringiter.Iterable) chan string {
	c := make(chan string)
	close(c)
	return c
}
