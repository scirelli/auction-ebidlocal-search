package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server"
	storefs "github.com/scirelli/auction-ebidlocal-search/internal/app/server/store/fs"
	ebidfsstore "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store/fs"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func main() {
	var logger = log.New("Main", log.DEFAULT_LOG_LEVEL)
	var configPath *string = flag.String("config-path", os.Getenv("SERVER_CONFIG"), "path to the config file.")
	var contentPath *string
	var appConfig *AppConfig
	var err error

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

	appConfig.Server.ContentPath = *contentPath
	fsStore := ebidfsstore.NewWatchlistStore(ebidfsstore.StoreConfig{
		WatchlistDir: appConfig.Server.WatchlistDir,
	}, logger)

	server.New(
		appConfig.Server,
		storefs.FSStore{
			storefs.NewUserStore(appConfig.Server.UserDir, appConfig.Server.DataFileName, logger),
			storefs.NewWatchlistStore(fsStore, logger),
		},
		log.New("Server", log.DEFAULT_LOG_LEVEL),
	).Run()
}
