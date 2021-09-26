package extract

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	libscrape "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/scrape"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

type AuctionItem struct {
	config  *Config
	logger  log.Logger
	scraper libscrape.HTMLScraper
}

func NewAuctionItem(config *Config) *AuctionItem {
	var logger = log.New("AuctionItemExtractor", config.LogLevel)
	removeImageSize := regexp.MustCompile(`(-[0-9]+x[0-9]+)(?P<extent>\..+$)`)
	extraWhiteSpace := regexp.MustCompile(`\s{2,}`)

	return &AuctionItem{
		config: config,
		logger: logger,
		scraper: libscrape.NewCompositeHTMLScrape(
			[]libscrape.HTMLScraper{
				//EndDate
				libscrape.ScrapeFunc(func(s *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
					var selector string = fmt.Sprintf("div.AuctionItem-listInfo input[name='%s']", "EndDate")
					var loc *time.Location
					var err error

					if loc, err = time.LoadLocation("America/New_York"); err != nil {
						return m
					}

					if v, exists := s.Find(selector).Attr("value"); exists {
						if m.EndDate, err = time.ParseInLocation("2021-12-07 6:45:00 PM", v, loc); err != nil {
							logger.Infof("'%s' could not parse EndDate", v)
						}
					} else {
						logger.Infof("'%s' does not exist", selector)
					}

					return m
				}),
				//"ImageURLs": "ImageUrls",
				libscrape.ScrapeFunc(func(s *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
					s.Find("div.carousel-inner img").Each(func(i int, selection *goquery.Selection) {
						if v, exists := selection.Attr("src"); exists {
							v = removeImageSize.ReplaceAllString(v, "${extent}")
							if u, err := url.Parse(v); err == nil {
								m.ImageURLs = append(m.ImageURLs, u)
							}
						}
					})

					return m
				}),
				//ItemURL
				libscrape.ScrapeFunc(func(s *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
					if v, err := url.Parse(s.Find("div.carousel-inner a.carousel-item").First().AttrOr("href", "")); err == nil {
						values := v.Query()
						values.Set("pageNumber", "pf6Q+hJtdeleDd9FfYpy9w==")
						v.RawQuery = values.Encode()
						m.ItemURL = v
					}
					return m
				}),
				//Extra Description
				libscrape.ScrapeFunc(func(s *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
					m.ExtendedDescription = strings.TrimSpace(extraWhiteSpace.ReplaceAllString(s.Find("div.tooltip-demos").First().Text(), " "))
					return m
				}),
			},
		).AddAll(
			//Inputs with simple string and int values
			func() []libscrape.HTMLScraper {
				var inputNamesTypeString = map[string]string{
					"Id":           "AuctionItemId",
					"ItemName":     "ItemName",
					"Types":        "Types",
					"SKUNumber":    "SKUNumber",
					"Description":  "Description",
					"StatusCode":   "StatusCode",
					"OriginalName": "OriginalName",
				}
				var inputNamesTypeInt = map[string]string{
					"TotalBids":            "TotalBids",
					"CurrentBidAmount":     "CurrentBidAmount",
					"MinimumNextBidAmount": "MinimumNextBidAmount",
					"BuyNowPrice":          "BuyNowPrice",
					"Quantity":             "Quantity",
					"ReservePrice":         "ReservePrice",
					"BidAmount":            "BidAmount",
				}
				var scrapers = make([]libscrape.HTMLScraper, 0, len(inputNamesTypeString)+len(inputNamesTypeInt)+2)

				for fieldName, inputName := range inputNamesTypeString {
					scrapers = append(scrapers, func(fieldName string, inputName string) libscrape.ScrapeFunc {
						return libscrape.ScrapeFunc(func(s *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
							var selector string = fmt.Sprintf("div.AuctionItem-listInfo input[name='%s']", inputName)
							if v, exists := s.Find(selector).Attr("value"); exists {
								reflect.ValueOf(m).Elem().FieldByName(fieldName).SetString(strings.TrimSpace(v))
							} else {
								logger.Infof("'%s' does not exist", selector)
							}

							return m
						})

					}(fieldName, inputName))
				}

				for fieldName, inputName := range inputNamesTypeInt {
					scrapers = append(scrapers, func(fieldName string, inputName string) libscrape.ScrapeFunc {
						return libscrape.ScrapeFunc(func(s *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
							var selector string = fmt.Sprintf("div.AuctionItem-listInfo input[name='%s']", inputName)
							if v, exists := s.Find(selector).Attr("value"); exists {
								i, _ := strconv.ParseInt(v, 10, 64)
								reflect.ValueOf(m).Elem().FieldByName(fieldName).SetInt(i)
							} else {
								logger.Infof("'%s' does not exist", selector)
							}

							return m
						})

					}(fieldName, inputName))
				}

				return scrapers
			}(),
		),
	}
}

//Extract implements the Extractor interface Extract. Extract expects a document of <div class="row"> top level elements. Rows should be the direct children of <body>
func (s *AuctionItem) Extract(in <-chan model.SearchResult) <-chan model.AuctionItem {
	var models = make(chan model.AuctionItem)

	go s.extract(in, models)

	return models
}

func (s *AuctionItem) extract(in <-chan model.SearchResult, out chan<- model.AuctionItem) {
	defer func() {
		close(out)
	}()

	for result := range in {
		doc, err := goquery.NewDocumentFromReader(ioutil.NopCloser(strings.NewReader(result.Content)))
		if err != nil {
			s.logger.Errorf("AuctionItemExtractor could not parse html from read stream '%s'", err)
			return
		}

		if os.Getenv("DEBUG") != "" {
			f, _ := ioutil.TempFile("/tmp", fmt.Sprintf("doc_%s_", "SearchResult"))
			d, _ := doc.Html()
			f.WriteString(d)
			f.Close()
		}

		doc.Find("body > div.row").Each(func(i int, selection *goquery.Selection) {
			m := model.AuctionItem{
				ParentAuctionID: result.AuctionID,
				Keywords:        []string{result.Keyword},
			}
			s.scraper.Scrape(selection, &m)

			if os.Getenv("DEBUG") != "" {
				if len(m.ImageURLs) == 0 {
					f, _ := ioutil.TempFile("/tmp", fmt.Sprintf("doc_%s_", "NoImageURLs"))
					d, _ := doc.Html()
					f.WriteString(d)
					f.Close()
				}
			}

			out <- m
		})
	}
}
