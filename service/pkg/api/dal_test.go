// Package api the API DAL test
package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-contrib/cache/persistence"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.uber.org/zap/zaptest"

	"github.com/aeekayy/stilla/service/pkg/models"
	// "github.com/aeekayy/stilla/service/lib/db"
)

// mockDB a mock database implemenation of DBIface
type mockDB struct {
	database		pgxmock.PgxPoolIface
	Lookup			map[string]string
}

// NewMockDB returns a new mock database 
func NewMockDB() (mockDB, error) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		return mockDB{}, err
	}

	m := make(map[string]string)

	return mockDB{
		database:	mock,
		Lookup:		m,
	}, nil
}

func (m mockDB) Close() {
	fmt.Println("closing mock postgresql database")
}

func (m mockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.database.Exec(ctx, sql, args)
}

func (m mockDB) Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error) {
	return m.database.Query(ctx, sql, optionsAndArgs)
}

func (m mockDB) QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row {
	return m.database.QueryRow(ctx, sql, optionsAndArgs)
}

func (m mockDB) GenerateAPIKey(name string, tags []string) (string, string, error) {
	hostID := uuid.New().String()
	apiKey := uuid.New().String()
	m.Lookup["ApiKey"] = apiKey
	m.Lookup["HostID"] = hostID
	m.Lookup["Hostname"] = name
	return hostID, apiKey, nil
}

func (m mockDB) ValidateAPIKey(id, token string) (string, error) {
	return m.Lookup["HostID"], nil
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
	pgDB, err := NewMockDB()
	if err != nil {
		t.Fatalf("could not create mock psql database: %s", err)
	}

	dal.Cache = testCache
	dal.Config = config
	dal.Logger = sugar
	dal.SessionKey = sessionKey
	dal.Collection = collection
	dal.CacheEnabled = true
	dal.Context = &ctx
	dal.Database = pgDB

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
		name                string
		configID            string
		hostID              string
		writeToCache        bool
		expectCacheHit      bool
		expectWriteCacheErr bool
		expectReadCacheErr  bool
	}{
		{"WriteToCachePass", "configTest", "testhost", true, true, false, false}, // tests read from cache too
		{"WriteToCachePassEmptyHost", "configTest", "", true, true, false, false},
		{"WriteToCacheFailEmptyConfig", "", "testHost", true, false, true, true},
		{"ReadFromCacheMiss", "configTest", "testhost", false, false, false, true},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			dal := setupDep(t)

			// write to the cache if we want to test this path
			if tc.writeToCache {
				err := dal.writeToCache(tc.configID, tc.hostID, basicBsonM)

				if tc.expectWriteCacheErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}

			// always read from the cache. We want to always test cache reads
			cacheHit, result, err := dal.readFromCache(tc.configID, tc.hostID)

			assert.Equal(t, tc.expectCacheHit, cacheHit, "expected the cache hit result to match.")

			// if there's an cache error,
			// make sure that it's not nil
			// if there's no cache error, make sure that
			// the err is nil and make sure that the
			// results match
			if tc.expectReadCacheErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, basicBsonM, result, "the cache values should match.")
			}
		})
	}
}
