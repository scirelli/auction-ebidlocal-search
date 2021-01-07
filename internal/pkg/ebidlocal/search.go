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
	SearchURL string = "https://auction.ebidlocal.com/cgi-bin/mmlist.cgi"
	//OpenAuctionsURL url that lists open auctions
	OpenAuctionsURL string = "https://www.ebidlocal.com/im-bidding/sale-events/"
	requestDelay           = 1
)

//Keywords list of keywords to search open auctions for.
type Keywords []string

//Search search all open auctions for the list of keywords.
func (kw Keywords) Search() chan string {
	var auctions []string = requestOpenAuctions(OpenAuctionsURL)
	var results = make(chan string, len(auctions))
	var offset time.Duration
	var wg sync.WaitGroup

	wg.Add(len(auctions))
	for _, auction := range auctions {
		go func(auction string, offset time.Duration) {
			defer wg.Done()
			time.Sleep(offset)
			log.Printf("Searching auction '%s'", auction)
			results <- searchAnAuction(auction, kw)
		}(auction, offset)
		offset += time.Second * requestDelay
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func searchAnAuction(auction string, keywords []string) (html string) {
	var res *http.Response
	var err error

	res, err = http.PostForm(SearchURL, url.Values{
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

func requestOpenAuctions(saleEventsURL string) []string {
	return getAuctions(scrapeAuctionUrls(saleEventsURL))
}

func scrapeAuctionUrls(saleEventsURL string) (urls []*url.URL) {
	res, err := http.Get(saleEventsURL)
	if err != nil {
		log.Println(err)
		return urls
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return urls
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return urls
	}

	doc.Find("div.widget_ebid_current_widget div.widgetOuter > a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}

		u, err := url.Parse(href)
		if err == nil {
			urls = append(urls, u)
		}
	})

	return urls
}

func getAuctions(us []*url.URL) (auctions []string) {
	auctions = make([]string, 0, len(us))
	for _, u := range us {
		if a := getAuction(u); a != "" {
			auctions = append(auctions, a)
		}
	}

	return
}

func getAuction(u *url.URL) string {
	return mapKeys(u.Query())[0]
}

func mapKeys(m map[string][]string) (keys []string) {
	keys = make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return
}
