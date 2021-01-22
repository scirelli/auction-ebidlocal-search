package ebidlocal

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	//SearchURL url to to append auction id to to search for items.
	SearchURL    string = "https://auction.ebidlocal.com/cgi-bin/mmlist.cgi"
	requestDelay        = 1
)

//Searchable interface for searchable keywords.
type Searchable interface {
	Search() <-chan string
}

var openAuctions *AuctionsCache

func init() {
	openAuctions = NewAuctionsCache()
}

//SearchFunc func that implements the Search interface.
type SearchFunc func() <-chan string

func (s SearchFunc) Search() <-chan string {
	return s()
}

func SearchAuctions(keywords []string, results chan string) {
	var offset time.Duration
	var wg sync.WaitGroup
	var auctions []string = openAuctions.GetAuctions()

	log.Printf("All auctions '%v'", auctions)

	wg.Add(len(auctions))
	for _, auction := range auctions {
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
}

func SearchAuction(auction string, keywords []string) (html string) {
	var res *http.Response
	var err error

	res, err = client.PostForm(SearchURL, url.Values{
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
