// Package db for database layer connection handling
package db

import (
	"context"
	"log"
	"testing"
	"time"
)

func setupDB(tb testing.TB) (*DBConn, func(tb testing.TB, d *DBConn), error) {
	log.Println("setup db")
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	dbUser := "postgres"
	dbPass := "postgres"
	dbHost := "localhost"
	dbName := "stilla"
	dbParams := ""

	pool, err := Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)

	if err != nil {
		log.Fatal("Can't connect to pool")
		return nil, nil, err
	}

	teardownFunc := func(tb testing.TB, pool *DBConn) {
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
	dbHost := "localhost"
	dbName := "stilla"
	dbParams := ""

	_, err := Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)

	if err != nil {
		t.Errorf("expected no error, received: %s", err)
	}
}

// TestSuiteAllQueries
func TestSuiteAllQueries(t *testing.T) {
	pgpool, teardownSuite, err := setupDB(t)
	defer teardownSuite(t, pgpool)

	if err != nil {
		t.Errorf("could not create database pool")
	}

	table := []struct {
		name          string
		inputName     string
		inputTags     []string
		errorExpected bool
	}{
		{"GenerateAPIKeyPass", "test", []string{"test1", "test2"}, false},
		{"GenerateAPIKeyFail", "", []string{"test1", "test2"}, true},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pgpool.GenerateAPIKey(tc.inputName, tc.inputTags)
			
			if (err != nil) != tc.errorExpected {
				t.Errorf("expected %t, got %s", tc.errorExpected, err)
			}
		})
	}
}
