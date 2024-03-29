package search

import (
	"strings"
	"sync"
	"time"
    "net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/iter/stringiter"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

/* NOTE:
List all current auctions: https://staples.prod4.maxanet.auction/Public/Auction/GetAuctions?filter=Current&pageSize=1000
	Each auctions URL: "#AuctionList div.product-dessc > div > a"

		Hidden input to get auction id: "input#hdn_AuctionId"

			Search URL for one auction: https://staples.prod4.maxanet.auction/Public/Auction/GetAuctionItems?AuctionId=74691&viewType=3&SearchFilter=nintendo&pageSize=100
*/

const (
	openAuctionsScheme        = "https"
	openAuctionsDomain string = "auction.ebidlocal.com"
	//openAuctionsPath url that lists open auctions
	openAuctionsPath       string        = "/Public/Auction/GetAuctions"
	openAuctionsQuery      string        = "?filter=Current&pageSize=1000"
	auctionRefreshInterval time.Duration = 10 * time.Minute
)

var clogger log.Logger

func init() {
	clogger = log.New("CachedAuctions", log.DEFAULT_LOG_LEVEL)
}

//NewAuctionsCache create auction cache instance.
func NewAuctionsCache() *AuctionsCache {
	return &AuctionsCache{
		refreshInterval: auctionRefreshInterval,
	}
}

//AuctionsCache stores the auctions cache.
type AuctionsCache struct {
	openAuctionCache []string
	lastRefresh      time.Time
	refreshInterval  time.Duration
	mux              sync.RWMutex
}

func (c *AuctionsCache) Iterator() stringiter.Iterator {
	return stringiter.SliceStringIterator(c.GetAuctions()).Iterator()
}

//RefreshAuctionCache refreshes the auctions cache.
func (c *AuctionsCache) RefreshAuctionCache() *AuctionsCache {
	var a []string = requestOpenAuctions(openAuctionsScheme + "://" + openAuctionsDomain + openAuctionsPath + openAuctionsQuery)
	c.mux.Lock()
	c.openAuctionCache = a
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
	return c.openAuctionCache
}

func requestOpenAuctions(openAuctionsURL string) []string {
	return getAuctionIds(scrapeAuctionLabelIds(openAuctionsURL))
}

func scrapeAuctionLabelIds(openAuctionsURL string) []string {
	var ids []string

    req, err := http.NewRequest("GET", openAuctionsURL, nil)
	if err != nil {
		clogger.Error(err)
		return ids
	}
    req.Header.Add("X-Requested-With", "XMLHttpRequest")
    req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	res, err := Client.Do(req)
	if err != nil {
		clogger.Error(err)
		return ids
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		clogger.Errorf("CachedAuctions.scrapeAuctionLabelIds: status code error: %d %s", res.StatusCode, res.Status)
		return ids
	}

	// f, err := os.Create("/tmp/dat2.html")
	// if err != nil {
	// 	return
	// }
	// defer f.Close()
	// _, err = io.Copy(f, res.Body)
	// clogger.Debug(err)

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		clogger.Error(err)
		return ids
	}

	doc.Find("span.label.label-warning").Each(func(i int, s *goquery.Selection) {
		id, ok := s.Attr("id")
		if !ok {
			return
		}

		ids = append(ids, id)
	})

	return ids
}

func getAuctionIds(labelIds []string) (auctionIds []string) {
	auctionIds = make([]string, 0, len(labelIds))

	for _, labelId := range labelIds {
		if id := getAuctionId(labelId); id != "" {
			auctionIds = append(auctionIds, id)
		}
	}

	return auctionIds
}

func getAuctionId(labelId string) string {
	parts := strings.Split(labelId, "_")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
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
