// Package db for database layer connection handling
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	dbPool    *pgxpool.Pool
	dbCtx     *context.Context
	mongoConn *mongo.Client
)

// APIKey ApiKey for authentication
type APIKey struct {
	ID      uuid.UUID `yaml:"id",json:"id"`
	Name    string    `yaml:"name",json:"name"`
	Role    uuid.UUID `yaml:"role",json:"role"`
	Created time.Time `yaml:"created",json:"created"`
	Updated time.Time `yaml:"updated",json:"updated"`
}

// APIKey API key for the application. This needs to move
// to another package. This does not belong in the db package.
type APIKey APIKey

// Connect connect to a Postgres compatible database.
func Connect(ctx *context.Context, dbUser, dbPass, dbHost, dbName, dbParams string) (*pgxpool.Pool, error) {
	// https://github.com/jackc/pgx/blob/master/batch_test.go#L32

	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s?%s", dbUser, dbPass, dbHost, dbName, dbParams)
	dbConfig, err := pgxpool.ParseConfig(connString)

	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %s", err)
	}

	pool, err := pgxpool.NewWithConfig(*ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %s", err)
	}

	dbCtx = ctx
	dbPool = pool
	return pool, nil
}

// MongoConnect establishes a connection to a MongoDB cluster
// https://www.geeksforgeeks.org/how-to-use-go-with-mongodb/
func MongoConnect(ctx *context.Context, dbUser, dbPass, dbHost, dbTimeout string) (*mongo.Client, context.Context, context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	mongoctx, cancel := context.WithTimeout(*ctx, 10*time.Second)

	uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority", dbUser, dbPass, dbHost)

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(mongoctx, options.Client().ApplyURI(uri))
	mongoConn = client
	return client, mongoctx, cancel, err
}

// GetAPIKey retrieves an API key from the database
// TODO: Add validation to the function
func GetAPIKey(keyID string) (APIKey, error) {
	var apiKey APIKey
	var apiID, apiName, apiRoleID string
	var apiCreated, apiUpdated time.Time

	err := dbPool.QueryRow(*dbCtx, "SELECT id, name, role, created, updated FROM api_keys WHERE id=$1;", keyID).Scan(&apiID, &apiName, &apiRoleID, &apiCreated, &apiUpdated)

	apiKey.ID = uuid.MustParse(apiID)
	apiKey.Name = apiName
	apiKey.Role = uuid.MustParse(apiRoleID)
	apiKey.Created = apiCreated
	apiKey.Updated = apiUpdated

	return apiKey, err
}

// GenerateAPIKey generate an api key and a public key for a new host
func GenerateAPIKey(name string, tags []string) (string, error) {
	var apiKeyID string

	err := dbPool.QueryRow(*dbCtx, "INSERT INTO api_keys(name, tags, private_key, salt, role) VALUES($1, $2, $3, $4, $5) RETURNING id;", name, tags, "", "", "e3f01984-8185-4829-affe-56b84a9913eb").Scan(&apiKeyID)

	return apiKeyID, err
}

// ValidateAPIKey validates an API Key for a host
func ValidateAPIKey(id string) (string, error) {
	var hostname string

	err := dbPool.QueryRow(*dbCtx, "SELECT name FROM api_keys WHERE id=$1;", id).Scan(&hostname)

	return hostname, err
}
