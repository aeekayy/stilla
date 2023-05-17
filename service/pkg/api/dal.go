package api

// "github.com/aeekayy/stilla/pkg/api/models"
import (
	"bytes"
	"context"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	pb "github.com/aeekayy/stilla/service/api/protobuf/messages"
	"github.com/aeekayy/stilla/service/lib/db"
	"github.com/aeekayy/stilla/service/pkg/api/models"
	svcmodels "github.com/aeekayy/stilla/service/pkg/models"
	"github.com/aeekayy/stilla/service/pkg/utils"
)

const (
	// TODO move this to configuration
	configDB                       = "configdb"
	configCollection               = "config"
	configVersionCollectionlection = "config_version"
	serviceName                    = "stilla"
	dateFormat                     = "2021-02-03T04:55:46.607+08:00"
)

// MongoQueryResult used to manage Mongo query results from channels
type MongoQueryResult struct {
	Result interface{} `json:"result"`
	Error  error       `json:"error"`
}

// DAL Data Access Layer struct for maintaining and managing
// data store connections for Stilla
type DAL struct {
	Cache         *persistence.RedisStore `json:"cache"`
	Config        *svcmodels.Config       `json:"config"`
	Database      *db.Conn                `json:"database"`
	Context       *context.Context        `json:"context"`
	DocumentStore *mongo.Client           `json:"document_store"`
	Logger        *zap.SugaredLogger      `json:"logger"`
	Producer      *kafka.Producer         `json:"producer"`
	APM           *newrelic.Application   `json:"apm"`
	Collection    string                  `json:"collection,omitempty"`
	SessionKey    string                  `json:"session_key"`
	CacheEnabled  bool                    `json:"cache_enabled"`
}

// AuditEvent audit event struct for sending messages of service events
type AuditEvent struct {
	message     interface{} `json:"message"`
	topic       string      `json:"topic"`
	messageType string      `json:"message_type"`
	function    string      `json:"function"`
}

// HostCache cache for host
type HostCache struct {
	Hostname string `json:"hostname"`
}

// NewDAL returns a new DAL
func NewDAL(ctx *context.Context, sugar *zap.SugaredLogger, apm *newrelic.Application, config *svcmodels.Config, dbConn *db.Conn, docStore *mongo.Client, cache *persistence.RedisStore, producer *kafka.Producer, collection, sessionKey string) *DAL {
	return &DAL{
		Context:       ctx,
		Config:        config,
		Database:      dbConn,
		DocumentStore: docStore,
		Cache:         cache,
		Collection:    collection,
		Logger:        sugar,
		Producer:      producer,
		SessionKey:    sessionKey,
		APM:           apm,
		CacheEnabled:  true, // default the cache to 'true' for now. TODO: Make this configurable.
	}
}

// ToByteSlice - converts AuditEvent into a byte slice
func (a *AuditEvent) ToByteSlice() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	err := enc.Encode(a)
	return buf.Bytes(), err
}

// RegisterHost registers a host and provides the requestor an API key
func (d *DAL) RegisterHost(ctx *gin.Context, hostRegisterIn models.HostRegisterIn, req interface{}) (string, string, error) {
	requestDetails := make(map[string]interface{})

	httpReq := req.(*http.Request)
	requestDetails["request.method"] = utils.SanitizeMessageValue(httpReq.Method)
	requestDetails["request.header"] = utils.SanitizeMessageValue(httpReq.Header)
	requestDetails["request.protocol"] = utils.SanitizeMessageValue(httpReq.Proto)
	requestDetails["request.contentlength"] = utils.SanitizeMessageValue(httpReq.ContentLength)
	requestDetails["request.host"] = utils.SanitizeMessageValue(httpReq.Host)
	requestDetails["request.uri"] = utils.SanitizeMessageValue(httpReq.RequestURI)
	requestDetails["request.remoteaddr"] = utils.SanitizeMessageValue(httpReq.RemoteAddr)
	requestDetails["host"] = utils.SanitizeMessageValue(hostRegisterIn)

	d.EmitMessage("config.audit", "HostRegister", requestDetails)

	hostID, apiKey, err := d.Database.GenerateAPIKey(hostRegisterIn.Name, hostRegisterIn.Tags)
	d.Logger.Infof("Generated API key")

	if err != nil {
		return "", "", fmt.Errorf("error registering host: %s", err)
	}

	return hostID, apiKey, err
}

// LoginHost uses the API Key of a host and validates it. Creates a new session if the key is valid
func (d *DAL) LoginHost(ctx *gin.Context, hostLoginIn models.HostLoginIn, req interface{}) (string, error) {
	requestDetails := make(map[string]interface{})

	httpReq := req.(*http.Request)
	requestDetails["request.method"] = utils.SanitizeMessageValue(httpReq.Method)
	requestDetails["request.header"] = utils.SanitizeMessageValue(httpReq.Header)
	requestDetails["request.protocol"] = utils.SanitizeMessageValue(httpReq.Proto)
	requestDetails["request.contentlength"] = utils.SanitizeMessageValue(httpReq.ContentLength)
	requestDetails["request.host"] = utils.SanitizeMessageValue(httpReq.Host)
	requestDetails["request.uri"] = utils.SanitizeMessageValue(httpReq.RequestURI)
	requestDetails["request.remoteaddr"] = utils.SanitizeMessageValue(httpReq.RemoteAddr)
	requestDetails["host"] = utils.SanitizeMessageValue(hostLoginIn)

	d.EmitMessage("config.audit", "HostLogin", requestDetails)

	hostKey, err := d.Database.ValidateAPIKey(hostLoginIn.APIKey, hostLoginIn.Host)

	if err != nil {
		return "", fmt.Errorf("invalid api key for host: %s", err)
	}

	d.Logger.Infof("Valid API Key")

	return hostKey, err
}

// InsertConfig insert a configuration object into the document store. This
// creates a new ConfigVersion object. The ObjectID of the ConfigVersion is then
// used to update the Config object reference for ConfigVersion
func (d *DAL) InsertConfig(ctx *gin.Context, configIn models.ConfigIn, req interface{}) (string, error) {
	requestDetails := make(map[string]interface{})

	httpReq := req.(*http.Request)
	requestDetails["request.method"] = utils.SanitizeMessageValue(httpReq.Method)
	requestDetails["request.header"] = utils.SanitizeMessageValue(httpReq.Header)
	requestDetails["request.protocol"] = utils.SanitizeMessageValue(httpReq.Proto)
	requestDetails["request.contentlength"] = utils.SanitizeMessageValue(httpReq.ContentLength)
	requestDetails["request.host"] = utils.SanitizeMessageValue(httpReq.Host)
	requestDetails["request.uri"] = utils.SanitizeMessageValue(httpReq.RequestURI)
	requestDetails["request.remoteaddr"] = utils.SanitizeMessageValue(httpReq.RemoteAddr)
	requestDetails["config"] = utils.SanitizeMessageValue(configIn)
	// get the host
	hostID := ctx.GetString("x-host-id")

	// select database and collection ith Client.Database method
	// and Database.Collection method
	d.EmitMessage("config.audit", "InsertConfig", requestDetails)

	configCollection := d.DocumentStore.Database(configDB).Collection(configCollection)
	configVersionCollection := d.DocumentStore.Database(configDB).Collection(configVersionCollectionlection)

	var result bson.M

	sanitizedConfigName := utils.SanitizeMongoInput(configIn.ConfigName)

	// the $where function is not support on the Atlas free tier
	// https://www.mongodb.com/docs/atlas/reference/free-shared-limitations/?_ga=2.189348331.1715576176.1677375251-1973124898.1674435602
	filter := bson.D{
		{Key: "config_name", Value: fmt.Sprintf("%s", sanitizedConfigName)},
	}

	opts := options.Update().SetUpsert(true)

	// see if there's an existing record
	err := configCollection.FindOne(
		ctx,
		filter,
	).Decode(&result)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err != mongo.ErrNoDocuments {
			return "", fmt.Errorf("error accessing the collection: %s", err)
		}
	}

	var configID string
	var version int32
	var created time.Time
	var wg sync.WaitGroup
	chConfigVersion := make(chan MongoQueryResult)
	chConfig := make(chan MongoQueryResult)

	if version = 1; result != nil {
		checkVersion := result["version"]
		if checkVersion != nil {
			version = checkVersion.(int32) + 1
		}

		checkConfigID := result["config_id"]

		if checkConfigID != nil {
			configID = checkConfigID.(string)
		}

		checkCreated := result["created"]

		if checkCreated != nil {
			created = checkCreated.(primitive.DateTime).Time()
		}
	}

	if configID == "" {
		configID = uuid.NewString()
	}

	updated := time.Now()
	checksum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s+%s:%s", configIn.ConfigName, configIn.Owner, created.String(), updated.String())))

	configVersionIn := bson.D{
		{"config", configIn.Config},
		{"checksum", fmt.Sprintf("%x", checksum)},
	}

	configAdd := bson.D{
		{"config_name", configIn.ConfigName},
		{"created_by", configIn.Owner},
		{"config", configVersionIn},
		{"config_id", configID},
		{"host", hostID},
		{"parents", configIn.Parents},
		{"created", created},
		{"modified", updated},
		{"version", version},
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		// create a new configVersion
		configVersion, err := configVersionCollection.InsertOne(ctx, configAdd)

		qr := MongoQueryResult{
			Result: configVersion,
			Error:  err,
		}

		chConfigVersion <- qr
	}()

	go func() {
		defer wg.Done()
		// UpdateOne accept two argument of type Context
		// and of empty interface
		updateDoc := bson.D{{"$set", configAdd}}
		config, err := configCollection.UpdateOne(ctx, filter, updateDoc, opts)

		qr := MongoQueryResult{
			Result: config,
			Error:  err,
		}

		chConfig <- qr
	}()

	qrConfigVersion := <-chConfigVersion

	if qrConfigVersion.Error != nil {
		d.Logger.Errorf("unable to insert configVersion: %v", qrConfigVersion.Error)
		return "", fmt.Errorf("unable to ingest configVersion object: %s", qrConfigVersion.Error)
	}

	qrConfig := <-chConfig

	if qrConfig.Error != nil {
		d.Logger.Errorf("unable to insert config: %v", qrConfig.Error)
		return "", fmt.Errorf("unable to insert config: %s", qrConfig.Error)
	}

	d.Logger.Infof("inserted configVersion %s", configID)
	d.Logger.Infof("inserted config object %s", configID)

	// write the configuration to the cache
	configResult := qrConfig.Result
	err = d.writeToCache(configID, hostID, configResult.(bson.M))

	return configID, err
}

// GetConfig returns a Config with the latest version of the ConfigVersion
func (d *DAL) GetConfig(ctx *gin.Context, configID string, hostID string, req interface{}) (models.ConfigResponse, error) {
	requestDetails := make(map[string]interface{})
	var configResponse models.ConfigResponse

	httpReq := req.(*http.Request)
	requestDetails["request.method"] = utils.SanitizeMessageValue(httpReq.Method)
	requestDetails["request.header"] = utils.SanitizeMessageValue(httpReq.Header)
	requestDetails["request.protocol"] = utils.SanitizeMessageValue(httpReq.Proto)
	requestDetails["request.contentlength"] = utils.SanitizeMessageValue(httpReq.ContentLength)
	requestDetails["request.host"] = utils.SanitizeMessageValue(httpReq.Host)
	requestDetails["request.uri"] = utils.SanitizeMessageValue(httpReq.RequestURI)
	requestDetails["request.remoteaddr"] = utils.SanitizeMessageValue(httpReq.RemoteAddr)

	// select database and collection ith Client.Database method
	// and Database.Collection method
	// check the cache first
	d.EmitMessage("config.audit", "GetConfig", requestDetails)

	configCollection := d.DocumentStore.Database(configDB).Collection(configCollection)
	cacheHit, err := d.readFromCache(configID, hostID)
	if err == nil {
		configResponse.Ingest(cacheHit)
		return configResponse, nil
	} else if err != persistence.ErrCacheMiss {
		return configResponse, err
	}

	var result bson.M
	var metadataKey string
	var metadataFilter bson.M
	var queryFilter []bson.M

	if primitive.IsValidObjectID(configID) {
		metadataKey = "_id"
		objID, err := primitive.ObjectIDFromHex(configID)

		if err != nil {
			return configResponse, fmt.Errorf("error setting objectid: %s, %s", configID, err)
		}
		metadataFilter = bson.M{"$eq": objID}
		queryFilter = append(queryFilter, bson.M{metadataKey: metadataFilter})
	} else {
		metadataKey = "config_name"
		metadataFilter = bson.M{"$eq": configID}
		queryFilter = append(queryFilter, bson.M{metadataKey: metadataFilter})
	}

	if hostID != "" {
		queryFilter = append(queryFilter, bson.M{"host": bson.M{"$eq": hostID}})
	}

	d.Logger.Debugf("config search filter: %v", bson.D{{"$and", queryFilter}})

	err = configCollection.FindOne(
		ctx,
		bson.D{{"$and", queryFilter}},
	).Decode(&result)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err == mongo.ErrNoDocuments {
			return configResponse, fmt.Errorf("the config document does not exist: %s", err)
		}

		return configResponse, fmt.Errorf("error accessing the config document: %s", err)
	}

	// TODO fix the response. The config version is empty
	configResponse.Ingest(&result)

	err = d.writeToCache(configID, hostID, result)

	return configResponse, err
}

// GetConfigs returns a paginated slice of Configs from the document store
func (d *DAL) GetConfigs(ctx *gin.Context, offset string, limit string, req interface{}) ([]models.ConfigStore, error) {
	requestDetails := make(map[string]interface{})

	httpReq := req.(*http.Request)
	requestDetails["limit"] = limit
	requestDetails["offset"] = offset
	requestDetails["request.method"] = utils.SanitizeMessageValue(httpReq.Method)
	requestDetails["request.header"] = utils.SanitizeMessageValue(httpReq.Header)
	requestDetails["request.protocol"] = utils.SanitizeMessageValue(httpReq.Proto)
	requestDetails["request.contentlength"] = utils.SanitizeMessageValue(httpReq.ContentLength)
	requestDetails["request.host"] = utils.SanitizeMessageValue(httpReq.Host)
	requestDetails["request.uri"] = utils.SanitizeMessageValue(httpReq.RequestURI)
	requestDetails["request.remoteaddr"] = utils.SanitizeMessageValue(httpReq.RemoteAddr)

	d.EmitMessage("config.audit", "GetConfigs", requestDetails)

	configCollection := d.DocumentStore.Database(configDB).Collection(configCollection)

	if limit == "" {
		limit = "100"
	}

	if offset == "" {
		offset = "0"
	}

	intLimit, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing the limit: %s", err)
	}
	intOffset, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing the offset: %s", err)
	}

	if intLimit > 100 {
		intLimit = 100
	}

	findOptions := options.Find()
	findOptions.SetLimit(intLimit)
	findOptions.SetSkip(intOffset)
	findOptions.SetProjection(bson.D{{"config_version", 0}})
	var results []models.ConfigStore

	// see if there's an existing record
	cursor, err := configCollection.Find(
		ctx,
		bson.D{{}},
	)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		return nil, fmt.Errorf("error accessing the documents: %s", err)
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("error accessing the cursor: %s", err)
	}

	return results, nil
}

// UpdateConfigByID Updates a configuration by the ID
func (d *DAL) UpdateConfigByID(ctx *gin.Context, configID string, updateConfigIn models.UpdateConfigIn, req interface{}) (interface{}, error) {
	requestDetails := make(map[string]interface{})

	httpReq := req.(*http.Request)
	requestDetails["request.method"] = utils.SanitizeMessageValue(httpReq.Method)
	requestDetails["request.header"] = utils.SanitizeMessageValue(httpReq.Header)
	requestDetails["request.protocol"] = utils.SanitizeMessageValue(httpReq.Proto)
	requestDetails["request.contentlength"] = utils.SanitizeMessageValue(httpReq.ContentLength)
	requestDetails["request.host"] = utils.SanitizeMessageValue(httpReq.Host)
	requestDetails["request.uri"] = utils.SanitizeMessageValue(httpReq.RequestURI)
	requestDetails["request.remoteaddr"] = utils.SanitizeMessageValue(httpReq.RemoteAddr)
	requestDetails["updateConfig"] = utils.SanitizeMessageValue(updateConfigIn)

	// select database and collection ith Client.Database method
	// and Database.Collection method
	d.EmitMessage("config.audit", "UpdateConfigByID", requestDetails)
	configCollection := d.DocumentStore.Database(configDB).Collection(configCollection)
	configVersionCollection := d.DocumentStore.Database(configDB).Collection(configVersionCollectionlection)

	var existingConfig bson.M

	sanitizedConfigID := utils.SanitizeMongoInput(configID)
	filter := bson.D{
		{Key: "_id", Value: sanitizedConfigID},
	}
	// see if there's an existing record
	err := configCollection.FindOne(
		ctx,
		filter,
	).Decode(&existingConfig)

	if err == mongo.ErrNoDocuments {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		return nil, fmt.Errorf("the document does not exist: %s", err)
	} else if err != nil {
		return nil, fmt.Errorf("error accessing the document: %s", err)
	}

	created := time.Now()
	updated := time.Now()
	checksum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s+%s:%s", updateConfigIn.ConfigName, updateConfigIn.Requester, created.String(), updated.String())))

	configVersionIn := bson.D{
		{"config", updateConfigIn.Config},
		{"config_name", updateConfigIn.ConfigName},
		{"checksum", checksum},
		{"created_by", updateConfigIn.Requester},
		{"created", created},
	}
	// create a new configVersion
	configVersion, err := configVersionCollection.InsertOne(ctx, configVersionIn)
	if err != nil {
		d.Logger.Errorf("unable to insert config_version: %v", err)
		return nil, fmt.Errorf("unable to ingest config_version object: %s", err)
	}

	d.Logger.Infof("Inserted config_version %v", configVersion.InsertedID)

	mapFilter := bson.M{"_id": sanitizedConfigID}

	// 6) Create the update
	update := bson.M{
		"$set": bson.M{"config_version": configVersion.InsertedID},
	}

	upsert := false
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// FindOneAndUpdate accept two argument of type Context
	// and of empty interface
	configResp := configCollection.FindOneAndUpdate(ctx, mapFilter, update, &opt)

	d.Logger.Infof("updated config object %s", configResp)
	return configResp, nil
}

// EmitMessage emits a message for the service. Currently only manages AuditEvents
func (d *DAL) EmitMessage(messageType, funcName string, body map[string]interface{}) {
	go func() {
		if d.Producer != nil {
			gob.Register(pb.AuditLog{})
			// convert body from map[string]interface{} to struct
			sbody, err := utils.MapToProtobufStruct(body)
			if err != nil {
				d.Logger.Errorf("error encoding message: %s", err)
				return
			}
			event := &pb.AuditLog{
				Message:     sbody,
				Topic:       messageType,
				MessageType: pb.AuditLog_AUDIT,
				FuncName:    funcName,
				Service:     serviceName,
			}

			// Write the new address book back to disk.
			out, err := proto.Marshal(event)

			if err != nil {
				d.Logger.Errorf("error encoding message: %s", err)
				return
			}

			//eventValue, _ := event.ToByteSlice()
			deliveryChan := make(chan kafka.Event, 10000)
			err = d.Producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &messageType, Partition: kafka.PartitionAny},
				Value:          out},
				deliveryChan,
			)

			if err != nil {
				d.Logger.Errorf("unable to emit event: %s", err)
			}
		}
	}()
}

// GetAuditLogs returns a pagination list of audit logs
func (d *DAL) GetAuditLogs(ctx *gin.Context, offset string, limit string, req interface{}) ([]pb.AuditLog, error) {
	requestDetails := make(map[string]interface{})

	httpReq := req.(*http.Request)
	requestDetails["limit"] = limit
	requestDetails["offset"] = offset
	requestDetails["request.method"] = utils.SanitizeMessageValue(httpReq.Method)
	requestDetails["request.header"] = utils.SanitizeMessageValue(httpReq.Header)
	requestDetails["request.protocol"] = utils.SanitizeMessageValue(httpReq.Proto)
	requestDetails["request.contentlength"] = utils.SanitizeMessageValue(httpReq.ContentLength)
	requestDetails["request.host"] = utils.SanitizeMessageValue(httpReq.Host)
	requestDetails["request.uri"] = utils.SanitizeMessageValue(httpReq.RequestURI)
	requestDetails["request.remoteaddr"] = utils.SanitizeMessageValue(httpReq.RemoteAddr)

	d.EmitMessage("config.audit", "GetAuditLogs", requestDetails)

	if limit == "" {
		limit = "100"
	}

	if offset == "" {
		offset = "0"
	}

	intLimit, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing the limit: %s", err)
	}
	intOffset, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing the offset: %s", err)
	}

	if intLimit > 100 {
		intLimit = 100
	}

	var results []pb.AuditLog

	rows, err := d.Database.Query(ctx, "SELECT id, service, funcname, body, created FROM audit ORDER BY created desc LIMIT $1 OFFSET $2;", intLimit, intOffset)

	if err != nil {
		return nil, fmt.Errorf("error retrieving audit logs: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r pb.AuditLog
		var msgID string
		err := rows.Scan(&msgID, &r.Service, &r.FuncName, &r.Message, &r.Sent)
		if err != nil {
			return nil, fmt.Errorf("error retrieving audit logs: %s", err)
		}
		d.Logger.Infof("adding to the message to result list: %s", msgID)
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error retrieving audit logs: %s", err)
	}

	return results, nil
}

// ValidateToken ...
func ValidateToken(dal *DAL, hostID string, token string) (string, bool, error) {
	var resp string

	resp, err := dal.Database.ValidateAPIKey(hostID, token)

	if resp != "" {
		return resp, true, nil
	}

	return "", false, fmt.Errorf("unable to retrieve host key: %s", err)
}

// readFromCache reads a configuration from the cache
func (d *DAL) readFromCache(configID, hostID string) (*bson.M, error) {
	var cacheValue *bson.M
	var respEnc string
	var hostPrefix string
	var resp bson.M

	// if the cache is not enabled. Skip all of this.
	if !d.CacheEnabled {
		// TODO: this should not be an error
		return cacheValue, fmt.Errorf("the cache is not enabled")
	}

	if hostID != "" {
		hostPrefix = fmt.Sprintf("_%s", hostID)
	}

	// set the cache key
	cacheKey := fmt.Sprintf("config_%s%s", configID, hostPrefix)
	err := d.Cache.Get(cacheKey, &respEnc)

	if err == nil {
		bsonBin, err := b64.StdEncoding.DecodeString(respEnc)
		if err != nil {
			d.Logger.Errorf("unable to retrieve config: %v", err)
			return cacheValue, fmt.Errorf("unable to retrieve config: %v", err)
		}
		err = bson.Unmarshal(bsonBin, resp)
		if err != nil {
			d.Logger.Errorf("unable to retrieve config: %v", err)
			return cacheValue, fmt.Errorf("unable to retrieve config: %v", err)
		}
	} else if err != persistence.ErrCacheMiss {
		d.Logger.Errorf("unable to retrieve config: %v", err)
		return cacheValue, fmt.Errorf("unable to retrieve config: %v", err)
	}

	logLine := utils.SanitizeLogMessageF("cache hit for %s", cacheKey)
	d.Logger.Info(logLine)
	return cacheValue, nil
}

// writeToCache writes a configuration to the cache
func (d *DAL) writeToCache(configID, hostID string, result bson.M) error {
	var hostPrefix string

	if hostID != "" {
		hostPrefix = fmt.Sprintf("_%s", hostID)
	}

	// set the cache key
	cacheKey := fmt.Sprintf("config_%s%s", configID, hostPrefix)

	logLine := utils.SanitizeLogMessage("setting cache for %s", cacheKey)
	d.Logger.Infof(logLine)
	bsonBin, err := bson.Marshal(result)
	if err != nil {
		d.Logger.Errorf("error writing to cache: %v", err)
		return fmt.Errorf("error writing to the cache %s", err)
	}
	cacheEnc := b64.StdEncoding.EncodeToString(bsonBin)
	err = d.Cache.Set(cacheKey, cacheEnc, time.Hour)
	if err != nil {
		d.Logger.Errorf("error writing to cache: %v", err)
		return fmt.Errorf("error writing to the cache %s", err)
	}

	return nil
}
