package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server"
	ebidLib "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/auctions"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func main() {
	var logger = log.New("Main")
	var configPath *string = flag.String("config-path", os.Getenv("EBIDLOCAL_CONFIG"), "path to the config file.")
	var contentPath *string
	var appConfig *AppConfig
	var err error

	cwd, err := os.Getwd()
	if err != nil {
		logger.Error.Fatalln(err)
	}
	logger.Info.Printf("Cwd '%s'\n", cwd)
	contentPath = flag.String("content-path", ".", fmt.Sprintf("Base path to save user and watchlist data. Default '%s'", "."))
	flag.Parse()

	logger.Info.Printf("config path '%s'\n", *configPath)
	if appConfig, err = LoadConfig(*configPath); err != nil {
		logger.Error.Fatalln(err)
	}

	appConfig.Ebidlocal.ContentPath = *contentPath
	appConfig.Server.ContentPath = *contentPath
	ebid := ebidlocal.New(appConfig.Ebidlocal)
	ebid.SetOpenAuctions(ebidLib.NewAuctionsCache())
	doneChan := make(chan struct{})
	go ebid.Scan(doneChan)
	server.New(appConfig.Server, ebid).Run()
	close(doneChan)
}
