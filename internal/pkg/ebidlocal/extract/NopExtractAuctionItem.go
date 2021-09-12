package extract

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

var NopExtractor, PassThroughExtractor = ExtractFunc(extract), ExtractFunc(extract)
var logger log.Logger = log.New("NopExtractor", log.DEFAULT_LOG_LEVEL)

func extract(in <-chan model.SearchResult) <-chan model.AuctionItem {
	var out = make(chan model.AuctionItem)
	close(out)
	go func() {
		logger.Info("Draining input channel")
		for item := range in {
			logger.Debugf("SKIPPING: '%s'", item)
		}
	}()
	return out
}
