package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/scanner"
	ebidLib "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/auctions"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func main() {
	var logger = log.New("Main")
	var configPath *string = flag.String("config-path", os.Getenv("SCANNER_CONFIG"), "path to the config file.")
	var contentPath *string
	var appConfig *AppConfig
	var err error
	ctx, cancel := context.WithCancel(context.Background())

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

	appConfig.Scanner.ContentPath = *contentPath

	scan := scanner.New(appConfig.Scanner)
	scan.SetOpenAuctions(ebidLib.NewAuctionsCache())
	scan.Scan(ctx)
	cancel()
}
