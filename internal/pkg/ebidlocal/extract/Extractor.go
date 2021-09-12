package extract

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
)

type Extractor interface {
	Extract(<-chan model.SearchResult) <-chan model.AuctionItem
}

type ExtractFunc func(<-chan model.SearchResult) <-chan model.AuctionItem

func (s ExtractFunc) Extract(in <-chan model.SearchResult) <-chan model.AuctionItem {
	return s(in)
}
