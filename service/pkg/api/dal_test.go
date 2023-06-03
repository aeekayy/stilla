// Package api the API DAL test
package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/gin-contrib/cache/persistence"

	"github.com/aeekayy/stilla/service/pkg/models"
)

// MockCache mocks the Redis cache for the DAL
type MockCache struct{
	mock.Mock
	persistence.RedisStore
}

type MockDAL struct{
	mock.Mock
	Cache		*MockCache
	Config		*models.Config
	Context		context.Context
	Collection	string
	SessionKey	string
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
func (m *MockCache) Replace(key string, data interface{}, expire time.Duration)  error {
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
func setupDep() *DAL {
	config := models.NewConfig()
	collection := "configdb"
	sessionKey := "testKey"

	ctx := context.Background()
	dal := new(DAL)
	testCache := new(MockCache)

	var response string
	cacheKey := "config_configID_hostID"
	testCache.On("Get", cacheKey, &response).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*string)
		resp := "test"
		arg = &resp
	})

	dal.Cache = testCache
	dal.Config = config
	dal.SessionKey = sessionKey
	dal.Collection = collection
	dal.Context = &ctx

	return dal
}

// TestReadFromCache validates a successful read from the cache
func TestReadFromCache(t *testing.T) {
	configID := "configID"
	hostID := "hostID"

	var expected *bson.M 

	dal := setupDep()

	bsonVal, err := dal.readFromCache(configID, hostID)

	// make sure err is nil
	assert.Nil(t, err)

	assert.Equal(t, expected, bsonVal, "the cache values should match.")
}