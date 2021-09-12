package scrape

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
)

type HTMLScraper interface {
	Scrape(*goquery.Selection, *model.AuctionItem) *model.AuctionItem
}

type ScrapeFunc func(*goquery.Selection, *model.AuctionItem) *model.AuctionItem

func (s ScrapeFunc) Scrape(selection *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
	return s(selection, m)
}

type CompositeHTMLScrape struct {
	scrapers []HTMLScraper
}

func NewCompositeHTMLScrape(s []HTMLScraper) *CompositeHTMLScrape {
	return &CompositeHTMLScrape{
		scrapers: s,
	}
}

func (s *CompositeHTMLScrape) Scrape(selection *goquery.Selection, m *model.AuctionItem) *model.AuctionItem {
	for _, scpr := range s.scrapers {
		m = scpr.Scrape(selection, m)
	}
	return m
}

func (s *CompositeHTMLScrape) Add(scraper HTMLScraper) *CompositeHTMLScrape {
	s.scrapers = append(s.scrapers, scraper)
	return s
}

func (s *CompositeHTMLScrape) AddAll(scrapers []HTMLScraper) *CompositeHTMLScrape {
	for _, scraper := range scrapers {
		s.Add(scraper)
	}
	return s
}
