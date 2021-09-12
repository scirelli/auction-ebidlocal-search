package store

import (
	"context"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
)

type NopWatchlistStore struct{}

func (s *NopWatchlistStore) SaveWatchlist(ctx context.Context, watchlist model.Watchlist) (string, error) {
	return "", nil
}
func (s *NopWatchlistStore) LoadWatchlist(ctx context.Context, watchlistID string) (model.Watchlist, error) {
	return model.Watchlist{}, nil
}
func (s *NopWatchlistStore) DeleteWatchlist(ctx context.Context, watchlistID string) error {
	return nil
}

type NopWatchlistContentStore struct{}

func (s *NopWatchlistContentStore) SaveWatchlistContent(ctx context.Context, watchlist *model.WatchlistContent) (string, error) {
	return "", nil
}
func (s *NopWatchlistContentStore) LoadWatchlistContent(ctx context.Context, watchlistID string) (*model.WatchlistContent, error) {
	return nil, nil
}
func (s *NopWatchlistContentStore) DeleteWatchlistContent(ctx context.Context, watchlistContentID string) error {
	return nil
}
