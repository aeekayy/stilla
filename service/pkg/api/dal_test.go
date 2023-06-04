// Package api the API DAL test
package api

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-contrib/cache/persistence"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.uber.org/zap/zaptest"

	"github.com/aeekayy/stilla/service/pkg/models"
)

// setupDep setup the dependencies for DAL testing
func setupDep(t *testing.T) *DAL {
	// logger for Zap
	logger := zaptest.NewLogger(t)
	sugar := logger.Sugar()

	// setup the configuration
	config := models.NewConfig()
	collection := "configdb"
	sessionKey := "testKey"

	ctx := context.Background()
	dal := new(DAL)
	// testCache := new(MockCache)

	s := miniredis.RunT(t)
	redisPwd := "redis"
	s.RequireAuth(redisPwd)
	config.Cache.Host = s.Addr()
	config.Cache.Username = "default"
	config.Cache.Password = redisPwd
	testCache := persistence.NewRedisCache(config.Cache.Host, config.Cache.Password, time.Second)

	// var response string
	// cacheKey := "config_configID_hostID"

	dal.Cache = testCache
	dal.Config = config
	dal.Logger = sugar
	dal.SessionKey = sessionKey
	dal.Collection = collection
	dal.CacheEnabled = true
	dal.Context = &ctx

	// add mongo
	// https://medium.com/@victor.neuret/mocking-the-official-mongo-golang-driver-5aad5b226a78

	return dal
}

// TestReadFromCacheDisabled validates an error when the cache is disabled
func TestReadFromCacheDisabled(t *testing.T) {
	configID := "configID"
	hostID := "hostID"

	dal := setupDep(t)
	dal.CacheEnabled = false

	_, _, err := dal.readFromCache(configID, hostID)

	// make sure err is not nil
	assert.NotNil(t, err)
}

// TestCache validates a successful write to the cache
func TestCache(t *testing.T) {
	basicBsonM := bson.M{"foo": "bar", "hello": "world"}

	table := []struct {
		name             string
		configID         string
		hostID           string
		writeToCache     bool
		writeCacheErr    bool
		readCacheErr     bool
		expectedCacheHit bool
	}{
		{"WriteToCachePass", "configTest", "testhost", true, false, false, true}, // tests read from cache too
		{"WriteToCachePassEmptyHost", "configTest", "", true, false, false, true},
		{"WriteToCacheFailEmptyConfig", "", "testHost", true, true, true, false},
		{"ReadFromCacheMiss", "configTest", "testhost", false, false, true, false},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			dal := setupDep(t)

			// write to the cache if we want to test this path
			if tc.writeToCache {
				err := dal.writeToCache(tc.configID, tc.hostID, basicBsonM)

				if tc.writeCacheErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}

			// always read from the cache. We want to always test cache reads
			cacheHit, result, err := dal.readFromCache(tc.configID, tc.hostID)

			assert.Equal(t, tc.expectedCacheHit, cacheHit, "expected the cache hit result to match.")

			// if there's an cache error, 
			// make sure that it's not nil 
			// if there's no cache error, make sure that
			// the err is nil and make sure that the 
			// results match
			if tc.readCacheErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, basicBsonM, result, "the cache values should match.")
			}
		})
	}
}
