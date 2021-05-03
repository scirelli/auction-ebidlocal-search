package scanner

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/publish"
)

//New constructor for scanner app. This app creates new watch lists on disk and has a scanner to keep them up-to-date.
func New(config Config) *Scanner {
	var logger = log.New("Scanner.New")

	return &Scanner{
		config:       config,
		logger:       logger,
		changePublsr: publish.NewStringChange(),
	}
}

//Scanner data for scanner ebidlocal app
type Scanner struct {
	config       Config
	logger       *log.Logger
	changePublsr publish.StringPublisher
}

func (s *Scanner) SubscribeForPath() (readChan <-chan string, unsubscribe func() error) {
	return s.changePublsr.Subscribe()
}

// Scan directory for watch lists and publishes the path. Use SubscribeForPath to be notified of found watch lists.
// Walk the watch list directory on an internval.
func (s *Scanner) Scan(ctx context.Context) error {
	timeBetweenRuns := time.Duration(s.config.ScanInterval) * time.Second
	watchlistDir := s.config.WatchlistDir

	s.logger.Info.Printf("Scanning '%s' at interval '%s'", watchlistDir, timeBetweenRuns)
	for {
		startTime := time.Now()

		if err := filepath.Walk(watchlistDir, s.walkCalback); err != nil {
			s.logger.Error.Printf("Error walking the path %q: %v\n", watchlistDir, err)
		}

		select {
		case <-ctx.Done():
			s.logger.Info.Println("Scanner stopped.")
			return nil
		default:
		}
		if elaspsedTime := time.Since(startTime); elaspsedTime < timeBetweenRuns {
			time.Sleep(timeBetweenRuns - elaspsedTime)
		}
	}
}

func (s *Scanner) walkCalback(path string, info os.FileInfo, err error) error {
	if err != nil {
		s.logger.Info.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
		return err
	}
	if info.Name() == s.config.DataFileName {
		s.logger.Info.Printf("Found file: %q\n", path)
		s.changePublsr.Publish(path)
	}

	return nil
}
