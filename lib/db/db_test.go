// Package db for database layer connection handling
package db

import (
	"context"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aeekayy/stilla/pkg/utils"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func setupDB(tb testing.TB) (*Conn, func(tb testing.TB, d *Conn), error) {
	log.Println("setup db")
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	dbUser := "postgres"
	dbPass := "postgres"
	dbHost := utils.GetEnv("POSTGRES_HOST", "localhost")
	dbName := "stilla"
	dbParams := "sslmode=disable"

	pool, err := Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)

	if err != nil {
		log.Fatal("Can't connect to pool")
		return nil, nil, err
	}

	teardownFunc := func(tb testing.TB, pool *Conn) {
		log.Println("teardown db")
		pool.Close()
	}

	return pool, teardownFunc, nil
}

// TestPassConnect
func TestPassConnect(t *testing.T) {
	ctx := context.Background()

	dbUser := "postgres"
	dbPass := "postgres"
	dbHost := utils.GetEnv("POSTGRES_HOST", "localhost")
	dbName := "stilla"
	dbParams := ""

	pool, err := Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)

	if err != nil {
		t.Errorf("expected no error, received: %s", err)
	}

	// validate the connection with a ping
	err = pool.ValidateConnection()

	if err != nil {
		t.Errorf("ping failed.")
	}
}

// TestValidateConnectionNoConnection tests ValidateConnection when no valid connection has been
// established
func TestValidateConnectionNoConnection(t *testing.T) {
	ctx := context.Background()

	dbUser := "postgres"
	dbPass := "postgres"
	dbHost := utils.GetEnv("POSTGRES_HOST", "localhost")
	dbName := "baddb"
	dbParams := ""

	pool, _ := Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)

	err := pool.ValidateConnection()

	if err == nil {
		t.Errorf("expected an error. received no error")
	}
}

// TestSuiteAllQueries
func TestSuiteAllQueries(t *testing.T) {
	pgpool, teardownSuite, err := setupDB(t)
	defer teardownSuite(t, pgpool)

	if err != nil {
		t.Errorf("could not create database pool")
	}

	if pgpool == nil {
		t.Errorf("expected a database connection")
	}

	table := []struct {
		name          string
		inputName     string
		inputTags     []string
		errorExpected bool
		checkKey      bool
	}{
		{"GenerateAPIKeyPass", "test", []string{"test1", "test2"}, false, false},
		{"GenerateAPIKeyFailEmptyName", "", []string{"test1", "test2"}, true, false},
		{"GenerateAPIKeyFailApiKey", "apikey", []string{"test1", "test2"}, true, false},
		{"GenerateAPIKeyFailName", "name", []string{"test1", "test2"}, true, false},
		{"GenerateAPIKeyPass", "test", []string{"test1", "test2"}, false, false},
		{"GenerateAPIKeyRandom", "{RANDOM}", []string{"random1"}, false, true},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			keyName := tc.inputName
			if keyName == "{RANDOM}" {
				keyName = randstring(10)
			}
			key, err := pgpool.GenerateAPIKey(keyName, tc.inputTags)

			if (err != nil) != tc.errorExpected {
				t.Errorf("expected %t, got the error %+v", tc.errorExpected, err)
			}

			if tc.checkKey {
				name, err := pgpool.ValidateAPIKey(key)

				if err != nil {
					t.Errorf("error validating the api key %s", key)
				}

				assert.Equal(t, name, keyName, "the two names of the api key should be the same.")
			}
		})
	}
}

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randstring(length int) string {
	return stringWithCharset(length, charset)
}
