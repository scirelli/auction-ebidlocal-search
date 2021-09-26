package filter

import (
	"strings"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	stringutils "github.com/scirelli/auction-ebidlocal-search/internal/pkg/stringUtils"
)

func ByKeyword(item model.AuctionItem) bool {
	keywordLookup := stringutils.SliceToDict(stringutils.ToLower(item.Keywords))
	for _, f := range stringutils.ToLower(stringutils.StripPunctuation(strings.Fields(item.String()))) {
		if _, exists := keywordLookup[f]; exists {
			return true
		}
	}
	return false
}
