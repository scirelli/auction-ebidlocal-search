package scanner

import (
	"context"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/publish"
)

//New constructor for scanner app. The scanner publishes watch lists it finds.
func New(config Config) *Scanner {
	changePublsr := publish.NewStringChange()
	changePublsr.PublishTTL = 10 * 60 * time.Second // The Scanner sets a long publish time because down stream handlers (watch list updater) an take a long time to process a list. Since the app was changed to process one list at a time (due to memory limitations) the publisher should give enough time for ebidlocal requests to finish.
	return &Scanner{
		config:       config,
		logger:       log.New("Scanner.New", log.DEFAULT_LOG_LEVEL),
		changePublsr: changePublsr,
	}
}

//Scanner data for scanner ebidlocal app
type Scanner struct {
	config       Config
	logger       log.Logger
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

	s.logger.Infof("Scanning '%s' at interval '%s'", watchlistDir, timeBetweenRuns)
	for {
		startTime := time.Now()

		if err := filepath.WalkDir(watchlistDir, s.walkCalback); err != nil {
			s.logger.Errorf("Error walking the path %q: %v\n", watchlistDir, err)
		}

		select {
		case <-ctx.Done():
			s.logger.Info("Scanner stopped.")
			return nil
		default:
		}
		if elaspsedTime := time.Since(startTime); elaspsedTime < timeBetweenRuns {
			time.Sleep(timeBetweenRuns - elaspsedTime)
		}
	}
}

func (s *Scanner) walkCalback(path string, d fs.DirEntry, err error) error {
	if err != nil {
		s.logger.Infof("prevent panic by handling failure accessing a path %q: %v\n", path, err)
		return err
	}
	if d.Name() == s.config.DataFileName {
		s.logger.Infof("Scan.walkCallback: Found file: %q\n", path)
		s.changePublsr.Publish(path)
	}

	return nil
}
