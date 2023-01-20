// Package db for database layer connection handling
package db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dbConn struct {
	dbUser string
	dbPass string
	dbHost string
	dbName string
}

// MockedDB is used in unit tests to mock db
func NewMockDB() pgxMock.PgxConnIface {
	return pgxmockConn.NewConn()
}

// TestPostgresConnect test Postgres connection function Connect
func TestPostgresConnect(t *testing.T) {
	t.Parallel()

	// the test matrix for Connect
	tests := []struct {
		name string
		in   dbConn
		out  string
	}{
		{"empty hostname", dbConn{"", "", "", ""}, "can't enter empty hostname name"},
	}

	for _, test := range tests {
		test := test
		t.Log(test.name)
		ctx := context.TODO()
		conn, err := Connect(ctx, test.in.dbUser, test.in.dbPass, test.in.dbHost, test.in.dbName)
		if err.Error != test.out {
			t.Error("%s test result mismatch. Got %s, exepcted %s", test.name, err, test.out)
		}
	}
}
