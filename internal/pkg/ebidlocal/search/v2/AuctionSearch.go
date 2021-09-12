package search //TODO: Rename package to v2

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/funcUtils"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	ebidhttp "github.com/scirelli/auction-ebidlocal-search/internal/pkg/net/http"
)

const (
	AuctionSite           string = "https://staples.prod4.maxanet.auction"
	SearchURL             string = AuctionSite + "/Public/Auction/GetAuctionItems"
	requestDelay                 = 1
	maxRetries                   = 3
	maxConcurrentRequests        = 5
)

var Client ebidhttp.HTTPClient
var throttle = funcUtils.ThrottleFuncFactory(maxConcurrentRequests)
var logger log.Logger

func init() {
	logger = log.New("Ebidlocal.Search", log.DEFAULT_LOG_LEVEL)
	Client = http.DefaultClient
}

func SearchAuctions(keywordIter stringiter.Iterable, openAuctions stringiter.Iterable) (results chan model.SearchResult) {
	results = make(chan model.SearchResult)

	go func() {
		var auctionIter stringiter.Iterator = openAuctions.Iterator()
		var wg sync.WaitGroup
		for auction, ok := auctionIter.Next(); ok; auction, ok = auctionIter.Next() {
			var kwIter stringiter.Iterator = keywordIter.Iterator()
			for keyword, ok := kwIter.Next(); ok; keyword, ok = kwIter.Next() {
				wg.Add(1)
				logger.Debugf("Searching '%s' for '%s'", auction, keyword)
				throttle(func(v ...interface{}) {
					defer wg.Done()
					var auction string = v[0].(string)
					var keyword string = v[1].(string)
					if err := SearchAuction(results, auction, keyword); err != nil {
						logger.Error(err)
					}
				}, auction, keyword)
			}
		}
		wg.Wait()
		logger.Debug("Completed Auction Searching")
		close(results)
	}()

	return results
}

func SearchAuction(out chan<- model.SearchResult, auction string, keyword string) (err error) {
	var res *http.Response
	var req *http.Request

	base, err := url.Parse(SearchURL)
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Add("AuctionId", auction)
	params.Add("SearchFilter", keyword)
	params.Add("viewType", "3")
	params.Add("pageSize", "10000")
	base.RawQuery = params.Encode()
	logger.Debugf("Making request to... URL '%s'; auction '%s'; keyword '%s'", base.String(), auction, keyword)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if req, err = http.NewRequestWithContext(ctx, "GET", base.String(), nil); err != nil {
		return err
	}
	req.Header.Add("Host", "staples.prod4.maxanet.auction")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36")
	req.Header.Add("Pragma", "no-cache")
	if res, err = Client.Do(req); err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Status code error: %d %s", res.StatusCode, res.Status))
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return err
	}
	removeDynamicData(fullyQualifyLinks(doc))
	if os.Getenv("DEBUG") != "" {
		f, _ := ioutil.TempFile("/tmp", fmt.Sprintf("doc_%s_", auction))
		d, _ := doc.Html()
		f.WriteString(d)
		f.Close()
	}

	doc.Find("div.wrapper-main div.ibox-content > div.row").Each(func(i int, s *goquery.Selection) {
		str, err := goquery.OuterHtml(s)
		if err != nil {
			return
		}
		out <- model.SearchResult{
			AuctionID: auction,
			Keyword:   keyword,
			Content:   str,
		}
	})

	return nil
}

func fullyQualifyLinks(doc *goquery.Document) *goquery.Document {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if partial, exists := s.Attr("href"); exists {
			s.SetAttr("href", AuctionSite+partial)
		}
	})
	return doc
}

func removeDynamicData(doc *goquery.Document) *goquery.Document {
	doc.Find(".product-timer.productimer-item.auction-timer").Remove()
	doc.Find("script").Remove()
	doc.Find("style").Remove()
	doc.Find("link").Remove()
	doc.Find("nav").Remove()
	return doc
}
