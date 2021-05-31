package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/notify"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/scanner"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/update"
	ebidLib "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/auctions"
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
	contentPath = flag.String("content-path", ".", fmt.Sprintf("Base path to save user and watchlist data. Default '%s'", "."))
	flag.Parse()

	logger.Infof("config path '%s'\n", *configPath)
	if appConfig, err = LoadConfig(*configPath); err != nil {
		logger.Fatal(err)
	}

	appConfig.Scanner.ContentPath = *contentPath
	appConfig.Updater.ContentPath = *contentPath

	//scanner produces paths
	scan := scanner.New(appConfig.Scanner)
	go scan.Scan(ctx)

	//Updater subscribes to the paths and checks for changes
	updater := update.New(ctx,
		storefs.FSStore{
			storefs.NewWatchlistStore(storefs.StoreConfig{
				WatchlistDir: appConfig.Updater.WatchlistDir,
				DataFileName: appConfig.Updater.DataFileName,
			}, log.New("Updater.FSStore", log.DEFAULT_LOG_LEVEL)),
		},
		appConfig.Updater)
	updater.SetOpenAuctions(ebidLib.NewAuctionsCache())
	pathsChan, _ := scan.SubscribeForPath()
	go updater.Update(pathsChan)

	//Any changes found are passed onto a notifier
	ch, _ := updater.SubscribeForChange()
	email := notify.EmailNotify{
		ServerUrl:    appConfig.Notifier.ServerUrl,
		Logger:       logger,
		WatchlistDir: appConfig.Notifier.WatchlistDir,
		MessageChan:  notify.NewWatchlistConvertData(appConfig.Notifier).Convert(ch),
	}
	email.Send()

	cancel()
}
