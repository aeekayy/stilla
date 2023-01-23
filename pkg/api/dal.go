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
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-contrib/cache/persistence"
	pgx "github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	pb "github.com/aeekayy/stilla/api/protobuf/messages"
	"github.com/aeekayy/stilla/lib/db"
	"github.com/aeekayy/stilla/pkg/api/models"
	"github.com/aeekayy/stilla/pkg/utils"
)

const (
	config_db                 = "configdb"
	config_collection         = "config"
	config_version_collection = "config_version"
	AUDIT_EVENT               = "audit"
	SERVICE_NAME              = "stilla"
)

// DAL Data Access Layer struct for maintaining and managing
// data store connections for Stilla
type DAL struct {
	Collection    string                  `json:"collection,omitempty"`
	Cache         *persistence.RedisStore `json:"cache"`
	Database      *pgx.Conn               `json:"database"`
	Context       *context.Context        `json:"context"`
	DocumentStore *mongo.Client           `json:"document_store"`
	Logger        *zap.SugaredLogger      `json:"logger"`
	Producer      *kafka.Producer         `json:"producer"`
}

// AuditEvent audit event struct for sending messages of service events
type AuditEvent struct {
	message     interface{} `json:"message"`
	topic       string      `json:"topic"`
	messageType string      `json:"message_type"`
	function    string      `json:"function"`
}

// NewDAL returns a new DAL
func NewDAL(ctx *context.Context, sugar *zap.SugaredLogger, dbConn *pgx.Conn, docStore *mongo.Client, cache *persistence.RedisStore, producer *kafka.Producer, collection string) *DAL {
	return &DAL{
		Context:       ctx,
		Database:      dbConn,
		DocumentStore: docStore,
		Cache:         cache,
		Collection:    collection,
		Logger:        sugar,
		Producer:      producer,
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
func (d *DAL) RegisterHost(ctx context.Context, hostRegisterIn models.HostRegisterIn, req interface{}) (string, error) {
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

	apiKey, err := db.GenerateAPIKey(hostRegisterIn.Name, hostRegisterIn.Tags)
	d.Logger.Infof("Generated API key")

	if err != nil {
		return "", fmt.Errorf("error registering host: %s", err)
	}

	return apiKey, err
}

// InsertConfig insert a configuration object into the document store. This
// creates a new ConfigVersion object. The ObjectID of the ConfigVersion is then
// used to update the Config object reference for ConfigVersion
func (d *DAL) InsertConfig(ctx context.Context, configIn models.ConfigIn, req interface{}) (interface{}, error) {
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

	// select database and collection ith Client.Database method
	// and Database.Collection method
	d.EmitMessage("config.audit", "InsertConfig", requestDetails)

	config_col := d.DocumentStore.Database(config_db).Collection(config_collection)
	config_version_col := d.DocumentStore.Database(config_db).Collection(config_version_collection)

	var result bson.M

	sanitizedConfigName := utils.SanitizeMongoInput(configIn.ConfigName)
	filter := bson.D{
		{Key: "$where", Value: bson.D{
				primitive.E{Key: "config_name", Value: sanitizedConfigName},
			},
		},
	}
	// see if there's an existing record
	err := config_col.FindOne(
		ctx,
		filter,
	).Decode(&result)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err != mongo.ErrNoDocuments {
			return nil, fmt.Errorf("error accessing the collection: %s", err)
		}
	} else {
		return nil, fmt.Errorf("the document exists")
	}

	created := time.Now()
	updated := time.Now()
	checksum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s+%s:%s", configIn.ConfigName, configIn.Owner, created.String(), updated.String())))

	config_version_in := bson.D{
		{"config", configIn.Config},
		{"config_name", configIn.ConfigName},
		{"checksum", checksum},
		{"created_by", configIn.Owner},
		{"created", created},
	}
	// create a new config_version
	config_version, err := config_version_col.InsertOne(ctx, config_version_in)
	if err != nil {
		d.Logger.Errorf("unable to insert config_version: %v", err)
		return nil, fmt.Errorf("unable to ingest config_version object: %s", err)
	}

	d.Logger.Infof("Inserted config_version %v", config_version.InsertedID)

	config_in := bson.D{
		{"config_name", configIn.ConfigName},
		{"created_by", configIn.Owner},
		{"config_version", config_version.InsertedID},
		{"parents", configIn.Parents},
		{"created", created},
		{"modified", updated},
	}

	// InsertOne accept two argument of type Context
	// and of empty interface
	config_resp, err := config_col.InsertOne(ctx, config_in)
	if err != nil {
		d.Logger.Errorf("unable to insert config: %v", err)
		return nil, fmt.Errorf("unable to insert config: %s", err)
	}

	d.Logger.Infof("created config object %s", config_resp.InsertedID)
	return config_resp.InsertedID, nil
}

// GetConfig returns a Config with the latest version of the ConfigVersion
func (d *DAL) GetConfig(ctx context.Context, configID string, req interface{}) (bson.M, error) {
	requestDetails := make(map[string]interface{})

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
	var resp bson.M
	var respEnc string
	cache_key := fmt.Sprintf("config_%s", configID)
	err := d.Cache.Get(cache_key, &respEnc)

	config_col := d.DocumentStore.Database(config_db).Collection(config_collection)
	config_version_col := d.DocumentStore.Database(config_db).Collection(config_version_collection)

	if err == nil {
		bsonBin, err := b64.StdEncoding.DecodeString(respEnc)
		if err != nil {
			d.Logger.Errorf("unable to retrieve config: %v", err)
			return nil, fmt.Errorf("unable to retrieve config: %v", err)
		}
		err = bson.Unmarshal(bsonBin, &resp)
		if err != nil {
			d.Logger.Errorf("unable to retrieve config: %v", err)
			return nil, fmt.Errorf("unable to retrieve config: %v", err)
		}
		logLine := utils.SanitizeLogMessage("cache hit for %s", cache_key)
		d.Logger.Infof(logLine)
		return resp, nil
	} else if err != persistence.ErrCacheMiss {
		d.Logger.Errorf("unable to retrieve config: %v", err)
		return nil, fmt.Errorf("unable to retrieve config: %v", err)
	}

	objID, err := primitive.ObjectIDFromHex(configID)

	if err != nil {
		return nil, fmt.Errorf("error setting objectid: %s", err)
	}

	var result bson.M
	// see if there's an existing record
	err = config_col.FindOne(
		ctx,
		bson.D{{"_id", bson.M{"$eq": objID}}},
	).Decode(&result)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("the document does not exist: %s", err)
		}

		return nil, fmt.Errorf("error accessing the document: %s", err)
	}

	var version_result bson.M

	err = config_version_col.FindOne(
		ctx,
		bson.D{{"_id", bson.M{"$eq": result["config_version"]}}},
		options.FindOne().SetProjection(bson.M{"_id": 0, "checksum.Subtype": 0}),
	).Decode(&version_result)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("the document does not exist: %s", err)
		}

		return nil, fmt.Errorf("error accessing the document: %s", err)
	}

	result["config"] = version_result
	logLine := utils.SanitizeLogMessage("setting cache for %s", cache_key)
	d.Logger.Infof(logLine)
	bsonBin, err := bson.Marshal(result)
	if err != nil {
		d.Logger.Errorf("error writing to cache: %v", err)
		return result, fmt.Errorf("error writing to the cache %s", err)
	}
	cacheEnc := b64.StdEncoding.EncodeToString(bsonBin)
	err = d.Cache.Set(cache_key, cacheEnc, time.Hour)
	if err != nil {
		d.Logger.Errorf("error writing to cache: %v", err)
		return result, fmt.Errorf("error writing to the cache %s", err)
	}

	return result, nil
}

// GetConfigs returns a paginated slice of Configs from the document store
func (d *DAL) GetConfigs(ctx context.Context, offset string, limit string, req interface{}) ([]models.ConfigStore, error) {
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

	config_col := d.DocumentStore.Database(config_db).Collection(config_collection)

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
	cursor, err := config_col.Find(
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
func (d *DAL) UpdateConfigByID(ctx context.Context, configID string, updateConfigIn models.UpdateConfigIn, req interface{}) (interface{}, error) {
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
	config_col := d.DocumentStore.Database(config_db).Collection(config_collection)
	config_version_col := d.DocumentStore.Database(config_db).Collection(config_version_collection)

	var existingConfig bson.M

	sanitizedConfigID := utils.SanitizeMongoInput(configID)
	filter := bson.D{
		{Key: "$where", Value: bson.D{
			primitive.E{Key: "_id", Value: sanitizedConfigID},
		},
		},
	}
	// see if there's an existing record
	err := config_col.FindOne(
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

	config_version_in := bson.D{
		{"config", updateConfigIn.Config},
		{"config_name", updateConfigIn.ConfigName},
		{"checksum", checksum},
		{"created_by", updateConfigIn.Requester},
		{"created", created},
	}
	// create a new config_version
	config_version, err := config_version_col.InsertOne(ctx, config_version_in)
	if err != nil {
		d.Logger.Errorf("unable to insert config_version: %v", err)
		return nil, fmt.Errorf("unable to ingest config_version object: %s", err)
	}

	d.Logger.Infof("Inserted config_version %v", config_version.InsertedID)

	mapFilter := bson.M{"_id": sanitizedConfigID}

	// 6) Create the update
	update := bson.M{
		"$set": bson.M{"config_version": config_version.InsertedID},
	}

	upsert := false
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// FindOneAndUpdate accept two argument of type Context
	// and of empty interface
	config_resp := config_col.FindOneAndUpdate(ctx, mapFilter, update, &opt)

	d.Logger.Infof("updated config object %s", config_resp)
	return config_resp, nil
}

// EmitMessage emits a message for the service. Currently only manages AuditEvents
func (d *DAL) EmitMessage(messageType, funcName string, body map[string]interface{}) {
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
			Service:     SERVICE_NAME,
		}

		// Write the new address book back to disk.
		out, err := proto.Marshal(event)

		if err != nil {
			d.Logger.Errorf("error encoding message: %s", err)
			return
		}

		//eventValue, _ := event.ToByteSlice()
		delivery_chan := make(chan kafka.Event, 10000)
		err = d.Producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &messageType, Partition: kafka.PartitionAny},
			Value:          out},
			delivery_chan,
		)

		if err != nil {
			d.Logger.Errorf("unable to emit event: %s", err)
		}
	}
}

// GetAuditLogs returns a pagination list of audit logs
func (d *DAL) GetAuditLogs(ctx context.Context, offset string, limit string, req interface{}) ([]pb.AuditLog, error) {
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
