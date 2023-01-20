// Package main initializes the Stilla service. This creates the logger, retrieves
// configuration and starts the web server
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"go.uber.org/ratelimit"
	"go.uber.org/zap"

	"github.com/aeekayy/stilla/pkg/api"
	"github.com/aeekayy/stilla/pkg/models"
)

const (
	defaultDomainName = "stilla.aeekay.co" // The default domain name
)

var (
	limit ratelimit.Limiter
	rps   = flag.Int("rps", 1000, "request per second")
)

func main() {
	ctx := context.Background()

	// start the logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Errorf("Error starting the logger. Exiting.")
		os.Exit(1)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Info("Starting Stilla")
	// get the configuration
	config, err := models.GetConfig()
	if err != nil {
		sugar.Errorf("There's an error retrieving the configuration: %s", err)
	}

	sugar.Infof("Retrieving variables for the environment %s", config.Environment)
	sugar.Infof("Using PostgreSQL database %s", config.Database.Name)

	// create the web server
	httpServer, err := api.Get(ctx, sugar, defaultDomainName, config)

	if err != nil {
		sugar.Errorf("Unable to start web server: %s", err)
	}

	sugar.Info("starting web server")
	go sugar.Fatal(httpServer.Run())
}
