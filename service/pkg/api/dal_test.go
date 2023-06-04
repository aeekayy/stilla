// Package api the API DAL test
package api

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-contrib/cache/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap/zaptest"

	"github.com/aeekayy/stilla/service/pkg/models"
)

// MockCache mocks the Redis cache for the DAL
type MockCache struct {
	mock.Mock
	persistence.RedisStore
}

type MockDAL struct {
	mock.Mock
	Cache      *MockCache
	Config     *models.Config
	Context    context.Context
	Collection string
	SessionKey string
}

// mocks the retrieval of a cache by the key cacheKey
// uses the key to set the string pointer response
func (m *MockCache) Get(key string, value interface{}) error {
	args := m.Called(key, value)

	return args.Error(0)
}

// Set sets an item to the cache, replacing any existing item.
func (m *MockCache) Set(key string, value interface{}, expire time.Duration) error {
	args := m.Called(key, value, expire)

	return args.Error(0)
}

// Add adds an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (m *MockCache) Add(key string, value interface{}, expire time.Duration) error {
	args := m.Called(key, value, expire)

	return args.Error(0)
}

// Replace sets a new value for the cache key only if it already exists. Returns an
// error if it does not.
func (m *MockCache) Replace(key string, data interface{}, expire time.Duration) error {
	args := m.Called(key, data, expire)

	return args.Error(0)
}

// Delete removes an item from the cache. Does nothing if the key is not in the cache.
func (m *MockCache) Delete(key string) error {
	args := m.Called(key)

	return args.Error(0)
}

// Increment increments a real number, and returns error if the value is not real
func (m *MockCache) Increment(key string, data uint64) (uint64, error) {
	args := m.Called(key, data)

	return uint64(args.Int(0)), args.Error(1)
}

// Decrement decrements a real number, and returns error if the value is not real
func (m *MockCache) Decrement(key string, data uint64) (uint64, error) {
	args := m.Called(key, data)

	return uint64(args.Int(0)), args.Error(1)
}

// Flush seletes all items from the cache.
func (m *MockCache) Flush() error {
	args := m.Called()

	return args.Error(0)
}

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

			if tc.writeToCache {
				err := dal.writeToCache(tc.configID, tc.hostID, basicBsonM)

				if tc.writeCacheErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}

			cacheHit, result, err := dal.readFromCache(tc.configID, tc.hostID)

			assert.Equal(t, tc.expectedCacheHit, cacheHit, "expected the cache hit result to match.")

			if tc.readCacheErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, basicBsonM, result, "the cache values should match.")
			}
		})
	}
}
