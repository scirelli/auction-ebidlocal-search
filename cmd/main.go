package main

import (
	"flag"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func main() {
	var logger = log.New("Main")
	var configPath *string = flag.String("config-path", os.Getenv("EBIDLOCAL_CONFIG"), "path to the config file.")
	var appConfig *AppConfig
	var err error

	flag.Parse()

	path, err := os.Getwd()
	if err != nil {
		logger.Error.Fatalln(err)
	}
	logger.Info.Printf("Cwd '%s'\n", path)

	logger.Info.Printf("config path '%s'\n", *configPath)
	if appConfig, err = LoadConfig(*configPath); err != nil {
		logger.Error.Fatalln(err)
	}

	var ebidlocal = ebidlocal.New(appConfig.Ebidlocal)
	var server = server.New(appConfig.Server, ebidlocal)
	server.Run()
}
