package ebidlocal

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	ebid "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
	ebidLib "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/generator"
)

const (
	//SearchURL url to to append auction id to to search for items.
	SearchURL    string = "https://auction.ebidlocal.com/cgi-bin/mmlist.cgi"
	requestDelay        = 1
	maxRetries          = 3
)

type AuctionSearcher interface {
	Search(keywords ebidLib.StringIterator, auctions ebidLib.StringIterator) (results chan string)
}

type AuctionSearchFunc func(keywords ebidLib.StringIterator, auctions ebidLib.StringIterator) (results chan string)

func (as AuctionSearchFunc) Search(keywords ebidLib.StringIterator, auctions ebidLib.StringIterator) (results chan string) {
	return as(keywords, auctions)
}

func SearchAuctions(keywordIter ebidLib.StringIterator, openAuctions ebidLib.StringIterator) (results chan string) {
	var offset time.Duration
	var wg sync.WaitGroup
	var keywords []string

	results = make(chan string)

	for keyword, done := keywordIter.Next(); done; keyword, done = keywordIter.Next() {
		keywords = append(keywords, keyword)
	}

	for auction, done := openAuctions.Next(); done; auction, done = openAuctions.Next() {
		wg.Add(1)
		go func(auction string, offset time.Duration) {
			defer wg.Done()
			time.Sleep(offset)
			log.Printf("Searching auction '%s'", auction)
			if html, err := SearchAuction(auction, keywords); err == nil {
				results <- html
			}
		}(auction, offset)
		offset += time.Second * requestDelay
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func SearchAuction(auction string, keywords []string) (html string, err error) {
	var res *http.Response

	res, err = ebid.Client.PostForm(SearchURL, url.Values{
		"auction": {auction},
		"keyword": {strings.Join(keywords, " ")},
		"stype":   {"ANY"},
		"search":  {"Go!"},
	})
	if err != nil {
		log.Println(err)
		return html, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return html, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return html, err
	}

	html, err = doc.Find("#DataTable tbody").First().Html()
	if err != nil {
		return html, err
	}

	log.Println("Got auction data.")
	return html, nil
}
