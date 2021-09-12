package model

type SearchResultAccessor interface {
	GetAuctionID() string
	GetKeyword() string
	GetContent() string
}

type SearchResult struct {
	AuctionID string
	Keyword   string
	Content   string
}

type AuctionIDKeywordSorter []SearchResult

func (a AuctionIDKeywordSorter) Len() int      { return len(a) }
func (a AuctionIDKeywordSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AuctionIDKeywordSorter) Less(i, j int) bool {
	return (a[i].AuctionID + a[i].Keyword) < (a[j].AuctionID + a[j].Keyword)
}
