package model

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/id"
)

type WatchlistContenter interface {
	id.IDer
	GetTimestamp() time.Time
	GetWatchlistID() string
	GetAuctionItems() []AuctionItem
}

type WatchlistContent struct {
	Id           string        `json:"id"`
	Timestamp    time.Time     `json:"timestamp"`
	WatchlistID  string        `json:"watchlistID"`
	AuctionItems []AuctionItem `json:"auctionItems"`
}

func (wc *WatchlistContent) ID() string {
	if wc.Id != "" {
		return wc.Id
	}

	var ids []string
	for _, item := range wc.AuctionItems {
		ids = append(ids, item.ID())
	}
	sort.Strings(ids)
	wc.Id = fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, ""))))
	return wc.Id
}

func (wc *WatchlistContent) GetTimestamp() time.Time {
	return wc.Timestamp
}

func (wc *WatchlistContent) GetWatchlistID() string {
	return wc.WatchlistID
}

func (wc *WatchlistContent) GetAuctionItems() []AuctionItem {
	return wc.AuctionItems
}
