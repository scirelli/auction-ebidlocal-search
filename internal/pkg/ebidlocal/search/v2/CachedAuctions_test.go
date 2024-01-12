package search

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/test/fixtures"
	"github.com/stretchr/testify/assert"
)

type CachedAuctionTestCase struct {
	Body       io.ReadCloser
	StatusCode int
	Error      error
	Expected   []string
}

func Skip_TestIntegration_scrapeAuctionUrls(t *testing.T) {
	u, _ := url.Parse(openAuctionsScheme + "://" + openAuctionsDomain + openAuctionsPath + openAuctionsQuery)
	var actual []string = scrapeAuctionLabelIds(u.String())

	if len(actual) == 0 {
		t.Error("No labels were scraped.")
	}
}

func Skip_TestIntegrationAuctionCache(t *testing.T) {
	var auctionCache *AuctionsCache = NewAuctionsCache()
	var auctions []string
	var iter = auctionCache.Iterator()

	for auction, done := iter.Next(); done; auction, done = iter.Next() {
		t.Log(auction)
		auctions = append(auctions, auction)
	}

	if len(auctions) == 0 {
		t.Error("No ideas found")
	}
}

func TestAuctionCache(t *testing.T) {
	var tests map[string]CachedAuctionTestCase = map[string]CachedAuctionTestCase{
		"Get auction ids from the warning label": CachedAuctionTestCase{
			Body:       fixtures.OpenFile(t, "../../../../../test/fixtures/internal/pkg/ebidlocal/search/v2/GetAuctions.html"),
			StatusCode: 200,
			Error:      nil,
			Expected:   []string{"75131", "74689", "74692", "74693", "103380", "74688", "74675", "74685", "74679", "74686", "74680", "74423"},
		},
		"GET returns an error": CachedAuctionTestCase{
			Body:       ioutil.NopCloser(strings.NewReader("hello world")),
			StatusCode: 200,
			Error:      errors.New("some error"),
			Expected:   []string{},
		},
		"GET gets a 404 response": CachedAuctionTestCase{
			Body:       ioutil.NopCloser(strings.NewReader("hello world")),
			StatusCode: 404,
			Error:      errors.New("some error 2"),
			Expected:   []string{},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			Client = &fixtures.MockClient{
				PostFormFunc: func(url string, data url.Values) (resp *http.Response, err error) {
					return nil, nil
				},
				GetFunc: func(url string) (resp *http.Response, err error) {
					return &http.Response{
						Body:       test.Body,
						StatusCode: test.StatusCode,
						Request:    &http.Request{},
					}, test.Error
				},
                DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       test.Body,
						StatusCode: test.StatusCode,
						Request:    &http.Request{},
					}, test.Error
                },
			}
			cache := AuctionsCache{
				refreshInterval: 0 * time.Minute,
			}
			iter := cache.Iterator()
			var result []string
			for auctionId, done := iter.Next(); done; auctionId, done = iter.Next() {
				result = append(result, auctionId)
			}
			assert.Equalf(t, len(test.Expected), len(result), "'%v' not equal '%v'", result, test.Expected)
			for i, v := range test.Expected {
				assert.Equalf(t, v, result[i], "'%v' not equal '%v'", result[i], v)
			}
		})
	}
}
