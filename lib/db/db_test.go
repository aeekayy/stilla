// Package db for database layer connection handling
package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbPool *pgxpool.Pool
)

type dbConn struct {
	dbUser string
	dbPass string
	dbHost string
	dbName string
}

// pgmock mocks a psql database
type pgmock struct {
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

// TestFailConnect
func TestFailConnect(t *testing.T) {
	ctx := context.Background()

	dbUser := ""
	dbPass := ""
	dbHost := ""
	dbName := ""
	dbParams := ""

	_, err := Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)

	if err == nil {
		t.Errorf("expected an error, received no error")
	}
}

func setup() {
	dbUser := "postgres"
	dbPass := "postgres"
	dbHost := "localhost"
	dbName := "stilla"
	dbParams := ""

	dbPool, _ := Connect(&ctx, dbUser, dbPass, dbHost, dbName, dbParams)
}
