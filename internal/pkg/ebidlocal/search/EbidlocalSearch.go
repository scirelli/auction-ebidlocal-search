package ebidlocal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"

	ebid "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/funcUtils"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

const (
	AuctionSite string = "https://auction.ebidlocal.com"
	//SearchURL url to to append auction id to to search for items.
	SearchURL             string = AuctionSite + "/cgi-bin/mmlist.cgi"
	requestDelay                 = 1
	maxRetries                   = 3
	maxConcurrentRequests        = 5
)

var throttle = funcUtils.ThrottleFuncFactory(maxConcurrentRequests)
var logger log.Logger

func init() {
	logger = log.New("Ebidlocal.Search", log.DEFAULT_LOG_LEVEL)
}

type AuctionSearcher interface {
	Search(keywords stringiter.Iterable, auctions stringiter.Iterable) (results chan string)
}

type AuctionSearchFunc func(keywords stringiter.Iterable, auctions stringiter.Iterable) (results chan string)

func (as AuctionSearchFunc) Search(keywords stringiter.Iterable, auctions stringiter.Iterable) (results chan string) {
	return as(keywords, auctions)
}

func SearchAuctions(keywordIter stringiter.Iterable, openAuctions stringiter.Iterable) (results chan string) {
	var keywords []string
	var iter stringiter.Iterator = keywordIter.Iterator()
	results = make(chan string)

	for keyword, done := iter.Next(); done; keyword, done = iter.Next() {
		keywords = append(keywords, keyword)
	}

	iter = openAuctions.Iterator()
	go func() {
		var wg sync.WaitGroup
		for auction, done := iter.Next(); done; auction, done = iter.Next() {
			wg.Add(1)
			throttle(func(v ...interface{}) {
				defer wg.Done()
				var auction string = v[0].(string)
				if html, err := SearchAuction(auction, keywords); err == nil {
					results <- html
				}
			}, auction)
		}
		wg.Wait()
		close(results)
	}()

	return results
}

func SearchAuction(auction string, keywords []string) (html string, err error) {
	var res *http.Response

	logger.Debugf("Searching... URL '%s'; auction '%s'; keywords '%s'", SearchURL, auction, keywords)
	res, err = ebid.Client.PostForm(SearchURL, url.Values{
		"auction": {auction},
		"keyword": {strings.Join(keywords, " ")},
		"stype":   {"ANY"},
		"search":  {"Go!"},
	})
	if err != nil {
		logger.Error(err)
		return html, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logger.Errorf("status code error: %d %s", res.StatusCode, res.Status)
		return html, err
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		logger.Error(err)
		return html, err
	}
	if os.Getenv("DEBUG") != "" {
		f, _ := ioutil.TempFile("/tmp", fmt.Sprintf("doc_%s_", auction))
		d, _ := doc.Html()
		f.WriteString(d)
		f.Close()
	}

	fullyQualifyLinks(doc)
	tbody := doc.Find("#DataTable tbody")

	html, err = tbody.First().Html()
	if err != nil {
		return html, err
	}

	return html, nil
}

func fullyQualifyLinks(doc *goquery.Document) *goquery.Document {
	doc.Find("#DataTable tbody tr td a").Each(func(i int, s *goquery.Selection) {
		if partial, exists := s.Attr("href"); exists {
			s.SetAttr("href", AuctionSite+partial)
		}
	})
	return doc
}

func removeDynamicData(doc *goquery.Document) *goquery.Document {
	doc.Find("#DataTable tbody td.highbidder span").Remove()
	doc.Find("#DataTable tbody td.currentamount span").Remove()
	doc.Find("#DataTable tbody td.nextbidrequired span, td.nextbidrequired a").Remove()
	doc.Find("#DataTable tbody td.yourbid span, td.yourbid input").Remove()
	doc.Find("#DataTable tbody td.yourmaximum span, td.yourmaximum input, td.yourmaximum br").Remove()
	return doc
}
