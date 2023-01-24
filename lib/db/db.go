// Package db for database layer connection handling
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	dbConn    *pgx.Conn
	dbCtx     *context.Context
	mongoConn *mongo.Client
)

// ApiKey ApiKey for authentication
type ApiKey struct {
	ID      uuid.UUID `yaml:"id",json:"id"`
	Name    string    `yaml:"name",json:"name"`
	Role    uuid.UUID `yaml:"role",json:"role"`
	Created time.Time `yaml:"created",json:"created"`
	Updated time.Time `yaml:"updated",json:"updated"`
}

// APIKey API key for the application. This needs to move
// to another package. This does not belong in the db package.
type APIKey ApiKey

// Connect connect to a Postgres compatible database.
func Connect(ctx *context.Context, dbUser, dbPass, dbHost, dbName, dbParams string) (*pgx.Conn, error) {
	// https://github.com/jackc/pgx/blob/master/batch_test.go#L32

	dsn := fmt.Sprintf("postgresql://%s:%s@%s/%s?%s", dbUser, dbPass, dbHost, dbName, dbParams)
	conn, err := pgx.Connect(*ctx, dsn)
	if err != nil {
		return nil, errors.New(fmt.Sprint("failed to connect database", err))
	}

	dbCtx = ctx
	dbConn = conn
	return conn, nil
}

// MongoConnect establishes a connection to a MongoDB cluster
// https://www.geeksforgeeks.org/how-to-use-go-with-mongodb/
func MongoConnect(ctx *context.Context, dbUser, dbPass, dbHost, dbTimeout string) (*mongo.Client, context.Context, context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	mongoctx, cancel := context.WithTimeout(*ctx, 30*time.Second)

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

	err := dbConn.QueryRow(*dbCtx, "SELECT id, name, role, created, updated FROM api_keys WHERE id=$1;", keyID).Scan(&apiID, &apiName, &apiRoleID, &apiCreated, &apiUpdated)

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

	err := dbConn.QueryRow(*dbCtx, "INSERT INTO api_keys(name, tags, private_key, salt, role) VALUES($1, $2, $3, $4, $5) RETURNING id;", name, tags, "", "", "e3f01984-8185-4829-affe-56b84a9913eb").Scan(&apiKeyID)

	return apiKeyID, err
}

// ValidateAPIKey validates an API Key for a host
func ValidateAPIKey(id string) (string, error) {
	var hostName string

	err := dbConn.QueryRow(*dbCtx, "SELECT name FROM api_keys WHERE id=$1;", id).Scan(&hostName)

	return hostName, err
}