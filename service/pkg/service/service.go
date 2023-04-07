// Package service initializes the Stilla service. This creates the logger, retrieves
// configuration and starts the web server
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"

	"github.com/aeekayy/stilla/service/pkg/api"
	"github.com/aeekayy/stilla/service/pkg/models"
)

const (
	defaultDomainName = "stilla.aeekay.co" // The default domain name
)

type Service struct {
	Name		string		`yaml:"name" json:"name"`
	ConfigFile	string		`yaml:"config_file" json:"config_file"`
	DomainName	string		`yaml:"domain_name" json:"domain_name"`
}

func NewService(configFile string) *Service {
	return &Service{
		Name: "stilla",
		ConfigFile: configFile,
		DomainName: defaultDomainName,
	}
}
// Start starts the service
func (s *Service) Start() error {
	ctx := context.Background()

	// start the logger
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("error starting the logger, exiting")
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Info("Starting Stilla")
	if s.ConfigFile != "" {
		sugar.Infof("Using the configuration file %s", s.ConfigFile)
	}
	// get the configuration
	config, err := models.GetConfig(s.ConfigFile)
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

	return nil
}
