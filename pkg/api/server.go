package api

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/go-session/gin-session"
	"github.com/pkg/errors"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"

	"github.com/aeekayy/stilla/lib/db"
	"github.com/aeekayy/stilla/pkg/models"
)

const (
	defaultHTTPPort       = 8080                       // The default web port. This should move to the configuration file
	defaultDomainName     = "stilla.aeekay.co"         // The default domain name
	defaultLongDomainName = "https://stilla.aeekay.co" // The full host with protocol
	letterBytes           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits         = 6                    // 6 bits to represent a letter index
	letterIdxMask         = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax          = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	uriStringCnt          = 8                    // The number of characters in the uri
	defaultHTTPTimeout    = 15
)

var (
	defaultReservedList = map[string]bool{
		"ping":  true,
		"error": true,
	}

	limit ratelimit.Limiter
	src   = rand.NewSource(time.Now().UnixNano())
)

// HTTPServer represents a web server
type HTTPServer struct {
	Context    context.Context
	DomainName string             `json:"domain_name",yaml:"domain_name"`
	Engine     *gin.Engine        `json:"engine",yaml:"engine"`
	Secure     bool               `json:"secure",yaml:"secure"`
	Logger     *zap.SugaredLogger `json:"logger",yaml:"logger"`
	server     *http.Server       `json:"server",yaml:"server"`
	config     models.Server      `json:"config",yaml:"config"`
}

// Get returns a new web server leveraging the service logger
func Get(ctx context.Context, sugar *zap.SugaredLogger, domainName string, config *models.Config) (*HTTPServer, error) {
	// add rate limiter
	limit = ratelimit.New(1000)

	// start the db connection
	dbUser := config.Database.Username
	dbPass := config.Database.Password
	dbHost := config.Database.Host
	dbName := config.Database.Name
	dbParams := config.Database.Parameters
	dbConn, err := db.Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)

	cachePass := config.Cache.Password
	cacheHost := config.Cache.Host

	store := persistence.NewRedisCache(cacheHost, cachePass, time.Second)

	if err != nil {
		sugar.Fatalf("couldn't connect to the database at %s: %w", dbHost, err)
		return nil, err
	}

	mongoConn, _, _, err := db.MongoConnect(&ctx, config.DocDB.Username, config.DocDB.Password, config.DocDB.Host, config.DocDB.Timeout)

	if err != nil {
		sugar.Fatalf("couldn't connect to the mongo database at %s: %w", config.DocDB.Host, err)
		return nil, err
	}

	var kafkaProducer *kafka.Producer
	if config.Audit {
		// Kafka producer
		kafkaProducer, err = kafka.NewProducer(config.GetKafkaConfig())
		sugar.Infof("%s", config.Kafka)

		if err != nil {
			sugar.Fatalf("failed to create producer: %s\n", err)
			return nil, err
		}
	}

	dal := NewDAL(&ctx, sugar, config, dbConn, mongoConn, store, kafkaProducer, "config", config.SessionKey)
	router := NewRouter(dal)

	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"https://stilla.aeekay.co"},
		AllowMethods:  []string{"PUT", "POST", "GET", "PATCH", "HEAD", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Referer", "Content-Type", "Accept", "Session", "Access-Control-Allow-Origin", "scheme", "path", "method", "authority", "user-agent", "sec-fetch-site", "sec-fetch-dest", "sec-fetch-mode", "sec-ch-ua-platform", "sec-ch-ua-mobile", "sec-ch-ua", "dnt", "content-length", "accept-encoding", "accept-language", "cache-control", "pragma"},
		ExposeHeaders: []string{"Origin", "Referer", "Content-Type", "Accept", "Session", "Access-Control-Allow-Origin", "scheme", "path", "method", "authority", "user-agent", "sec-fetch-site", "sec-fetch-dest", "sec-fetch-mode", "sec-ch-ua-platform", "sec-ch-ua-mobile", "sec-ch-ua", "dnt", "content-length", "accept-encoding", "accept-language", "cache-control", "pragma"},
		AllowOriginFunc: func(origin string) bool {
			// origins := []string{"http://localhost:8080", "https://api.aeekay.co", "http://api.aeekay.co:8080", "https://poseidon.aeekay.co", "https://ehgosolutions.com"}
			if origin == "http://localhost:8080" {
				return true
			}

			if origin == "https://stilla.aeekay.co" {
				return true
			}

			return false
		},
		MaxAge: 12 * time.Hour,
	}))

	// use session for API Keys
	router.Use(ginsession.New())

	addr := fmt.Sprintf(":%d", config.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &HTTPServer{Context: ctx, Engine: router, DomainName: domainName, server: srv, config: config.Server}, nil
}

// run runs the web server
func (h *HTTPServer) run() error {
	if h.Secure {
		return autotls.RunWithContext(h.Context, h.Engine, h.DomainName)
	}
	
	return h.Engine.Run()
}

// Run runs the web server
func (h *HTTPServer) Run() error {
	timeoutDuration, err := time.ParseDuration(h.config.Timeout)
	if err != nil {
		return fmt.Errorf("error parsing the duration: %s", err)
	}

	quit := make(chan error)

	go func() {
		if err := h.run(); err != nil {
			quit <- err
		}
	}()

	// SIGKILL and SIGSTOP cannot be caught, so don't bother adding them here
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-interrupt:
		fmt.Println("Caught interrupt, gracefully shutting down")
	case err := <-quit:
		if err != http.ErrServerClosed {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	return errors.Wrap(h.server.Shutdown(ctx), "Failed shutting down gracefully")
}
