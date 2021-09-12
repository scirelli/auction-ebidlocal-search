package update

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/extract"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search"
)

type SearchExtractor interface {
	extract.Extractor
	search.AuctionSearcher
}

type EbidlocalExtractor struct {
	extract.Extractor
	search.AuctionSearcher
}
