package ebidlocal

import (
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	ebid "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
)

const (
	//openAuctionsURL url that lists open auctions
	openAuctionsURL        string        = "https://www.ebidlocal.com/im-bidding/sale-events/"
	auctionRefreshInterval time.Duration = 10 * time.Minute
)

//NewAuctionsCache create auction cache instance.
func NewAuctionsCache() *AuctionsCache {
	return &AuctionsCache{
		refreshInterval: auctionRefreshInterval,
	}
}

//AuctionsCache stores the auctions cache.
type AuctionsCache struct {
	auctions        []string
	lastRefresh     time.Time
	refreshInterval time.Duration
	mux             sync.RWMutex
}

func (c *AuctionsCache) Iterator() stringiter.Iterator {
	return stringiter.SliceStringIterator(c.GetAuctions()).Iterator()
}

//RefreshAuctionCache refreshes the auctions cache.
func (c *AuctionsCache) RefreshAuctionCache() *AuctionsCache {
	var a []string = requestOpenAuctions(openAuctionsURL)
	c.mux.Lock()
	c.auctions = a
	c.lastRefresh = time.Now()
	defer c.mux.Unlock()
	return c
}

//GetAuctions retrieve the cached auctions.
func (c *AuctionsCache) GetAuctions() []string {
	if time.Since(c.lastRefresh) > c.refreshInterval {
		c.RefreshAuctionCache()
	}
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.auctions
}

func requestOpenAuctions(saleEventsURL string) []string {
	return getAuctions(scrapeAuctionUrls(saleEventsURL))
}

func scrapeAuctionUrls(saleEventsURL string) (urls []*url.URL) {
	res, err := ebid.Client.Get(saleEventsURL)
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
