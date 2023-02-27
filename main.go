// Package main initializes the Stilla service. This creates the logger, retrieves
// configuration and starts the web server
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"

	"github.com/aeekayy/stilla/pkg/api"
	"github.com/aeekayy/stilla/pkg/models"
)

const (
	defaultDomainName = "stilla.aeekay.co" // The default domain name
)

var (
	limit      ratelimit.Limiter
	rps        = flag.Int("rps", 1000, "request per second")
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "", "Configuration file for Stilla")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	// start the logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Errorf("error starting the logger, exiting")
		os.Exit(1)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Info("Starting Stilla")
	if configFile != "" {
		sugar.Infof("Using the configuration file %s", configFile)
	}
	// get the configuration
	config, err := models.GetConfig(configFile)
	if err != nil {
		sugar.Errorf("There's an error retrieving the configuration: %s", err)
	}

	// enable tracing if it's enable
	if config.Sentry.Enabled {
		sugar.Info("Starting Sentry")
		err := sentry.Init(sentry.ClientOptions{
			// Either set your DSN here or set the SENTRY_DSN environment variable.
			Dsn: config.Sentry.DSN,
			BeforeSendTransaction: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				// Here you can inspect/modify transaction events before they are sent.
				// Returning nil drops the event.
				if strings.Contains(event.Message, "test-transaction") {
					// Drop the transaction
					return nil
				}
				return event
			},
			// Enable tracing
			EnableTracing: true,
			// Specify either a TracesSampleRate...
			TracesSampleRate: 1.0,
			// ... or a TracesSampler
			TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
				// As an example, this custom sampler does not send some
				// transactions to Sentry based on their name.
				hub := sentry.GetHubFromContext(ctx.Span.Context())
				name := hub.Scope().Transaction()
				if name == "GET /favicon.ico" {
					return 0.0
				}
				if strings.HasPrefix(name, "HEAD") {
					return 0.0
				}
				// As an example, sample some transactions with a uniform rate.
				if strings.HasPrefix(name, "POST") {
					return 0.2
				}
				// Sample all other transactions for testing. On
				// production, use TracesSampleRate with a rate adequate
				// for your traffic, or use the SamplingContext to
				// customize sampling per-transaction.
				return 1.0
			}),
		})

		if err != nil {
			sugar.Errorf("Unable to start up Sentry: %s", err)
		}

		// Flush buffered events before the program terminates.
		// Set the timeout to the maximum duration the program can afford to wait.
		defer sentry.Flush(2 * time.Second)
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
