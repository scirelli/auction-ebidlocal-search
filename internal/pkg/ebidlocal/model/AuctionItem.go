package model

import (
	"fmt"
	"net/url"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/id"
)

type AuctionItemAccessor interface {
	id.IDer
	SearchableDescription() string

	GetParentAuctionID() string
	GetName() string
	GetDescription() string
	GetBidAmount() int
	GetNextMinBidAmout() int
	GetImages() []*url.URL
	GetItemURL() *url.URL
	GetKeywords() []string
}

//AuctionItem currently matches data from Ebidlocal v2 site
type AuctionItem struct {
	Id                   string     `json:"id"`
	ParentAuctionID      string     `json:"parentAuctionId"`
	ImageURLs            []*url.URL `json:imageUrls,omitempty"`
	ItemURL              *url.URL   `json:"itemUrl,omitempty"`
	TotalBids            int        `json:"totalBids,omitempty"`
	CurrentBidAmount     int        `json:"currentBidAmount,omitempty"`
	ItemName             string     `json:"itemName,omitempty"`
	Keywords             []string   `json:"keywords,omitempty"`
	MinimumNextBidAmount int        `json:"minimumNextBidAmount,omitempty"`
	BuyNowPrice          int        `json:"buyNowPrice,omitempty"`
	Quantity             int        `json:"quantity,omitempty"`
	Types                string     `json:"types,omitempty"`
	SKUNumber            string     `json:"skuNumber,omitempty"`
	Description          string     `json:"description,omitempty"`
	ExtendedDescription  string     `json:"extendedDescription,omitempty"`
	EndDate              time.Time  `json:"endDate,omitempty"`
	StatusCode           string     `json:"statusCode,omitempty"`
	ReservePrice         int        `json:"reservePrice,omitempty"`
	BidAmount            int        `json:"bidAmount,omitempty"`
	OriginalName         string     `json:"originalName,omitempty"`
}

func (a *AuctionItem) String() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", a.Id, a.ItemName, a.Description, a.ExtendedDescription, a.OriginalName)
}

func (a *AuctionItem) ID() string {
	return a.Id
}

func (a *AuctionItem) SearchableDescription() string {
	return a.String()
}

func (a *AuctionItem) GetName() string {
	return a.ItemName
}

func (a *AuctionItem) GetDescription() string {
	return a.Description
}

func (a *AuctionItem) GetBidAmount() int {
	return a.BidAmount
}

func (a *AuctionItem) GetNextMinBidAmout() int {
	return a.MinimumNextBidAmount
}

func (a *AuctionItem) GetImages() []*url.URL {
	return a.ImageURLs
}

func (a *AuctionItem) GetItemURL() *url.URL {
	return a.ItemURL
}

func (a *AuctionItem) GetKeywords() []string {
	return a.Keywords
}

//---------------- Sorting -------------------------
//ByID implements the sort.Sort interface to sort the Models by it's IDer
type ByID []AuctionItemAccessor

func (s ByID) Len() int {
	return len(s)
}
func (s ByID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByID) Less(i, j int) bool {
	return s[i].ID() < s[j].ID()
}

//------------------ Grouping -------------------------
//AuctionItemGroupByKeyword groups the AuctionItems by keyword, returns a map of kwyword to AuctionItems
type AuctionItemGroupByKeyword []AuctionItem

func (g AuctionItemGroupByKeyword) Group() map[string][]AuctionItem {
	var set = make(map[string][]AuctionItem)
	for _, item := range g {
		for _, keyword := range item.Keywords {
			if _, exists := set[keyword]; exists {
				set[keyword] = append(set[keyword], item)
			} else {
				set[keyword] = []AuctionItem{item}
			}
		}
	}
	return set
}

//----------------- Filters ---------------------------
//Filter a function with accepts and AuctionItem and returns true if it should be allowed to pass the filter.
type Filter func(AuctionItem) bool

//AuctionItemChanFilterer a single method interface which creates a new channel with all elements that pass the test implemented by the provided Filter function. This filter function should close the output channel.
type AuctionItemChanFilterer interface {
	Filter(Filter) <-chan AuctionItem
}

type FilterAuctionItemChan <-chan AuctionItem

func (f FilterAuctionItemChan) Filter(filter Filter) <-chan AuctionItem {
	var out = make(chan AuctionItem)

	go func() {
		defer close(out)
		for item := range f {
			if filter(item) {
				out <- item
			}
		}
	}()

	return out
}

//AuctionItemSliceFilterer a single method interface which creates a new slice with all elements that pass the test implemented by the provided Filter function.
type AuctionItemSliceFilterer interface {
	Filter(Filter) []AuctionItem
}

type FilterAuctionItemSlice []AuctionItem

func (f FilterAuctionItemSlice) Filter(filter Filter) (out []AuctionItem) {
	out = make([]AuctionItem, 0, len(f)>>1)
	for _, item := range f {
		if filter(item) {
			out = append(out, item)
		}
	}

	return
}
