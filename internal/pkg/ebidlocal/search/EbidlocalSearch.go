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
)

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
			results <- SearchAuction(auction, keywords)
		}(auction, offset)
		offset += time.Second * requestDelay
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func SearchAuction(auction string, keywords []string) (html string) {
	var res *http.Response
	var err error

	res, err = ebid.Client.PostForm(SearchURL, url.Values{
		"auction": {auction},
		"keyword": {strings.Join(keywords, " ")},
		"stype":   {"ANY"},
		"search":  {"Go!"},
	})
	if err != nil {
		log.Println(err)
		return html
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return html
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return html
	}

	html, _ = doc.Find("#DataTable tbody").First().Html()
	log.Println("Got auction data.")
	return html
}
