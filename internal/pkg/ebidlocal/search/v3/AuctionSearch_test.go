package search

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/test/fixtures"
	"github.com/stretchr/testify/assert"
)

type AuctionSearchTestCase struct {
	Responses  []*http.Response
	StatusCode int
	Error      error
	Auctions   stringiter.Iterable
	Keywords   stringiter.Iterable
	Expected   []model.SearchResult
}

func TestSearchAuction(t *testing.T) {
	var tests map[string]AuctionSearchTestCase = map[string]AuctionSearchTestCase{
		"Should return all rows of the Ebidlocal search results for one auction and one keyword": AuctionSearchTestCase{
			Responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
			},
			Error:    nil,
			Auctions: stringiter.SliceStringIterator([]string{"auction1"}),
			Keywords: stringiter.SliceStringIterator([]string{"hi"}),
			Expected: []model.SearchResult{
				{AuctionID: "auction1", Keyword: "hi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "hi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
			},
		},
		"Should return all rows of the Ebidlocal search results for one auction and two keywords": AuctionSearchTestCase{
			Responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
			},
			Error:    nil,
			Auctions: stringiter.SliceStringIterator([]string{"auction1"}),
			Keywords: stringiter.SliceStringIterator([]string{"shi", "thanos"}),
			Expected: []model.SearchResult{
				{AuctionID: "auction1", Keyword: "thanos", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "thanos", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
			},
		},
		"Should return all rows of the Ebidlocal search results for two auctions and one keyword": AuctionSearchTestCase{
			Responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
			},
			Error:    nil,
			Auctions: stringiter.SliceStringIterator([]string{"auction1", "auction2"}),
			Keywords: stringiter.SliceStringIterator([]string{"shi"}),
			Expected: []model.SearchResult{
				{AuctionID: "auction2", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction2", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
			},
		},
		"Should return all rows of the Ebidlocal search results for two auctions and two keywords": AuctionSearchTestCase{
			Responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
				&http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
					<html>
					<head>
					  <title></title>
					</head>
					<body>
					  <div class="wrapper-main mb-3">
						<div class="ibox-content border">
						  <div class="row pb-3 mt-2 border-bottom"></div>
						  <div class="row pb-3 mt-2 border-bottom"></div>
						</div>
					  </div>
					</body>
					</html>`)),
					StatusCode: 200,
					Request:    &http.Request{},
				},
			},
			Error:    nil,
			Auctions: stringiter.SliceStringIterator([]string{"auction1", "auction2"}),
			Keywords: stringiter.SliceStringIterator([]string{"shi", "thanos"}),
			Expected: []model.SearchResult{
				{AuctionID: "auction1", Keyword: "thanos", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "thanos", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction2", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction2", Keyword: "thanos", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction1", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction2", Keyword: "shi", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
				{AuctionID: "auction2", Keyword: "thanos", Content: `<div class="row pb-3 mt-2 border-bottom"></div>`},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			var responseCount int = -1
			Client = &fixtures.MockClient{
				PostFormFunc: func(url string, data url.Values) (resp *http.Response, err error) {
					return nil, nil
				},
				GetFunc: func(url string) (resp *http.Response, err error) {
					return nil, nil
				},
				DoFunc: func(req *http.Request) (resp *http.Response, err error) {
					responseCount++
					return test.Responses[responseCount], test.Error
				},
			}
			var results []model.SearchResult

			resultsChan := SearchAuctions(test.Keywords, test.Auctions)
			for item := range resultsChan {
				results = append(results, item)
			}
			sort.Sort(model.AuctionIDKeywordSorter(test.Expected))
			sort.Sort(model.AuctionIDKeywordSorter(results))
			for i, expected := range test.Expected {
				assert.Equalf(t, expected, results[i], "'%v' not equal '%v'", results[i], expected)
			}
		})
	}
}
