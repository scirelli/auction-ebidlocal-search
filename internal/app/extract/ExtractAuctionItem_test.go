package extract

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/stretchr/testify/assert"
)

type Test struct {
	Doc      string
	Expected []model.AuctionItem
}

var simpleInput = `
		<div class="row">
			<div class="AuctionItem-listInfo">
				<input name="%s" value="%s"/>
			</div>
		</div>`

func toURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}

var tests = map[string]Test{
	"Should extract AuctionItem Id": {
		Doc: fmt.Sprintf(simpleInput, "AuctionItemId", "11810030"),
		Expected: []model.AuctionItem{
			{
				Id: "11810030",
			},
		},
	},
	"Should extract ItemName": {
		Doc: fmt.Sprintf(simpleInput, "ItemName", "a name"),
		Expected: []model.AuctionItem{
			{
				ItemName: "a name",
			},
		},
	},
	"Should extract Types": {
		Doc: fmt.Sprintf(simpleInput, "Types", "a type"),
		Expected: []model.AuctionItem{
			{
				Types: "a type",
			},
		},
	},
	"Should extract SKUNumber": {
		Doc: fmt.Sprintf(simpleInput, "SKUNumber", "a sku number"),
		Expected: []model.AuctionItem{
			{
				SKUNumber: "a sku number",
			},
		},
	},
	"Should extract Description": {
		Doc: fmt.Sprintf(simpleInput, "Description", "This is a description"),
		Expected: []model.AuctionItem{
			{
				Description: "This is a description",
			},
		},
	},
	"Should extract StatusCode": {
		Doc: fmt.Sprintf(simpleInput, "StatusCode", "status 1"),
		Expected: []model.AuctionItem{
			{
				StatusCode: "status 1",
			},
		},
	},
	"Should extract OriginalName": {
		Doc: fmt.Sprintf(simpleInput, "OriginalName", "original name"),
		Expected: []model.AuctionItem{
			{
				OriginalName: "original name",
			},
		},
	},
	"Should extract TotalBids": {
		Doc: fmt.Sprintf(simpleInput, "TotalBids", "1"),
		Expected: []model.AuctionItem{
			{
				TotalBids: 1,
			},
		},
	},
	"Should extract CurrentBidAmount": {
		Doc: fmt.Sprintf(simpleInput, "CurrentBidAmount", "10"),
		Expected: []model.AuctionItem{
			{
				CurrentBidAmount: 10,
			},
		},
	},
	"Should extract MinimumNextBidAmount": {
		Doc: fmt.Sprintf(simpleInput, "MinimumNextBidAmount", "11"),
		Expected: []model.AuctionItem{
			{
				MinimumNextBidAmount: 11,
			},
		},
	},
	"Should extract BuyNowPrice": {
		Doc: fmt.Sprintf(simpleInput, "BuyNowPrice", "12"),
		Expected: []model.AuctionItem{
			{
				BuyNowPrice: 12,
			},
		},
	},
	"Should extract Quantity": {
		Doc: fmt.Sprintf(simpleInput, "Quantity", "13"),
		Expected: []model.AuctionItem{
			{
				Quantity: 13,
			},
		},
	},
	"Should extract ReservePrice": {
		Doc: fmt.Sprintf(simpleInput, "ReservePrice", "14"),
		Expected: []model.AuctionItem{
			{
				ReservePrice: 14,
			},
		},
	},
	"Should extract BidAmount": {
		Doc: fmt.Sprintf(simpleInput, "BidAmount", "15"),
		Expected: []model.AuctionItem{
			{
				BidAmount: 15,
			},
		},
	},
	"Should extract ImageURLs": {
		Doc: `
		<div class="row">
			<div class="carousel-inner">
				<a href="#"><img src="1"></a>
				<a href="#"><img src="2"></a>
				<a href="#"><img src="3"></a>
				<a href="#"><img src="4"></a>
			</div>
			<div class="AuctionItem-listInfo">
				<input name="%s" value="%s"/>
			</div>
		</div>`,
		Expected: []model.AuctionItem{
			{
				ImageURLs: []*url.URL{toURL("1"), toURL("2"), toURL("3"), toURL("4")},
			},
		},
	},
	"Should not parse sub-rows": {
		Doc: `
		<div class="row">
			<div class="carousel-inner">
				<a href="#"><img src="1"></a>
				<a href="#"><img src="2"></a>
				<a href="#"><img src="3"></a>
				<a href="#"><img src="4"></a>
			</div>
			<div class="row">
				<div class="AuctionItem-listInfo">
					<input name="%s" value="%s"/>
				</div>
			</div>
		</div>`,
		Expected: []model.AuctionItem{
			{
				ImageURLs: []*url.URL{toURL("1"), toURL("2"), toURL("3"), toURL("4")},
			},
		},
	},
	"Should parse multiple rows": {
		Doc: `
		<div class="row">
			<div class="carousel-inner">
				<a href="#"><img src="1"></a>
				<a href="#"><img src="2"></a>
				<a href="#"><img src="3"></a>
				<a href="#"><img src="4"></a>
			</div>
			<div class="row">
				<div class="AuctionItem-listInfo">
					<input name="%s" value="%s"/>
				</div>
			</div>
		</div>
		<div class="row">
			<div class="carousel-inner">
				<a href="#"><img src="5"></a>
				<a href="#"><img src="6"></a>
				<a href="#"><img src="7"></a>
				<a href="#"><img src="8"></a>
			</div>
			<div class="row">
				<div class="AuctionItem-listInfo">
					<input name="%s" value="%s"/>
				</div>
			</div>
		</div>`,
		Expected: []model.AuctionItem{
			{
				ImageURLs: []*url.URL{toURL("1"), toURL("2"), toURL("3"), toURL("4")},
			},
			{
				ImageURLs: []*url.URL{toURL("5"), toURL("6"), toURL("7"), toURL("8")},
			},
		},
	},
}

func Test_AuctionItem(t *testing.T) {
	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			results := make([]model.AuctionItem, 0, len(test.Expected))
			extractor := NewAuctionItem(&Config{})
			in := make(chan model.SearchResult, 1)
			in <- model.SearchResult{
				Content: test.Doc,
			}
			close(in)
			for m := range extractor.Extract(in) {
				results = append(results, m)
			}

			assert.Equalf(t, len(test.Expected), len(results), "Results differ")
			for i, m := range test.Expected {
				assert.Equalf(t, m.Id, results[i].Id, "Ids are not equal '%s' != '%s'", m.Id, results[i].Id)
				assert.Equal(t, m.TotalBids, results[i].TotalBids, "TotalBids are not equal '%d' != '%d'", m.TotalBids, results[i].TotalBids)
				assert.Equal(t, m.CurrentBidAmount, results[i].CurrentBidAmount, "CurrentBidAmount are not equal '%d' != '%d'", m.CurrentBidAmount, results[i].CurrentBidAmount)
				assert.Equal(t, m.ItemName, results[i].ItemName, "ItemName are not equal '%s' != '%s'", m.ItemName, results[i].ItemName)
				assert.Equal(t, m.MinimumNextBidAmount, results[i].MinimumNextBidAmount, "MinimumNextBidAmount are not equal '%d' != '%d'", m.MinimumNextBidAmount, results[i].MinimumNextBidAmount)
				assert.Equal(t, m.BuyNowPrice, results[i].BuyNowPrice, "BuyNowPrice are not equal '%d' != '%d'", m.BuyNowPrice, results[i].BuyNowPrice)
				assert.Equal(t, m.Quantity, results[i].Quantity, "Quantity are not equal '%d' != '%d'", m.Quantity, results[i].Quantity)
				assert.Equal(t, m.Types, results[i].Types, "Types are not equal '%s' != '%s'", m.Types, results[i].Types)
				assert.Equal(t, m.SKUNumber, results[i].SKUNumber, "SKUNumber are not equal '%s' != '%s'", m.SKUNumber, results[i].SKUNumber)
				assert.Equal(t, m.Description, results[i].Description, "Description are not equal '%s' != '%s'", m.Description, results[i].Description)
				assert.Equal(t, m.EndDate, results[i].EndDate, "EndDate are not equal '%s' != '%s'", m.EndDate, results[i].EndDate)
				assert.Equal(t, m.StatusCode, results[i].StatusCode, "StatusCode are not equal '%s' != '%s'", m.StatusCode, results[i].StatusCode)
				assert.Equal(t, m.ReservePrice, results[i].ReservePrice, "ReservePrice are not equal '%d' != '%d'", m.ReservePrice, results[i].ReservePrice)
				assert.Equal(t, m.BidAmount, results[i].BidAmount, "BidAmount are not equal '%d' != '%d'", m.BidAmount, results[i].BidAmount)
				assert.Equal(t, m.OriginalName, results[i].OriginalName, "OriginalName are not equal '%s' != '%s'", m.OriginalName, results[i].OriginalName)
				assert.Equal(t, m.ImageURLs, results[i].ImageURLs, "ImageURLs are not equal '%v' != '%v'", m.ImageURLs, results[i].ImageURLs)
			}
		})
	}
}

func Skip_Test_Integration_AuctionItem(t *testing.T) {
	extractor := NewAuctionItem(&Config{})
	retrievedItems := false
	for item := range extractor.Extract(ebidlocal.AuctionSearchFactory("v2", nil).Search(stringiter.SliceStringIterator([]string{"car"}))) {
		retrievedItems = true
		t.Logf("AuctionItem '%s'\n\n", item.String())
		assert.NotEmpty(t, item.ImageURLs, "Image URLs should not be empty. This is a test created from an error where sub-rows were being scraped.")
	}
	assert.True(t, retrievedItems, "No items were extracted")
}
