package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/notify"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/scanner"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/update"
	search "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/search"
	storefs "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store/fs"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func main() {
	var logger = log.New("Scanner.Main", log.DEFAULT_LOG_LEVEL)
	var configPath *string = flag.String("config-path", os.Getenv("SCANNER_CONFIG"), "path to the config file.")
	var contentPath *string
	var appConfig *AppConfig
	var err error
	ctx, cancel := context.WithCancel(context.Background())

	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Cwd '%s'\n", cwd)
	contentPath = flag.String("content-path", "", fmt.Sprintf("Base path to save user and watchlist data. Default '%s'", "."))
	flag.Parse()

	logger.Infof("config path '%s'\n", *configPath)
	if appConfig, err = LoadConfig(*configPath); err != nil {
		logger.Fatal(err)
	}

	logger.LogLevel = log.GetLevel(appConfig.LogLevel)
	logger.Infof("Log level is set to: '%s'", logger.LogLevel)

	if *contentPath != "" {
		appConfig.Scanner.ContentPath = *contentPath
		appConfig.Updater.ContentPath = *contentPath
	}

	//scanner produces paths
	scan := scanner.New(appConfig.Scanner)

	//Updater subscribes to the paths and checks for changes
	updater := update.New(ctx,
		storefs.FSStore{
			storefs.NewWatchlistStore(storefs.StoreConfig{
				WatchlistDir: appConfig.Updater.WatchlistDir,
				DataFileName: appConfig.Updater.DataFileName,
			}, log.New("Updater.FSStore", appConfig.Scanner.LogLevel)),
		},
		search.AuctionSearchFactory("v1", nil),
		appConfig.Updater)
	pathsChan, _ := scan.SubscribeForPath()

	//Any changes found are passed onto a notifier
	watchlistChangeEvent, _ := updater.SubscribeForChange()
	email := notify.EmailNotify{
		ServerUrl:    appConfig.Notifier.ServerUrl,
		Logger:       logger,
		WatchlistDir: appConfig.Notifier.WatchlistDir,
		MessageChan: notify.NewFilter(func(msg notify.NotificationMessage) bool {
			return msg.User.Verified
		}).Filter(ctx, notify.NewDedupeQueue().Enqueue(notify.NewWatchlistConvertData(appConfig.Notifier).Convert(watchlistChangeEvent))),
	}

	go scan.Scan(ctx)
	go updater.Update(pathsChan)
	email.Send()

	cancel()
}
