/*
 * Buffet Config Manager
 *
 * A configuration service that stores and retrieves configuration.
 *
 * API version: 0.1.0
 * Contact: apiteam@swagger.io
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ConfigStore ...
type ConfigStore struct {
	ID primitive.ObjectID `json:"id,omitempty",bson:"_id"`
	// Unique name for the configuration
	ConfigName    string               `json:"config_name",bson:"config_name"`
	Owner         string               `json:"owner",bson:"owner"`
	ConfigVersion ConfigVersion        `json:"config_version",bson:"config_version"`
	Parents       []primitive.ObjectID `json:"parents,omitempty",bson:"parents,omitempty"`
	Created       time.Time            `json:"created,omitempty",bson:"created,omitempty"`
	Modified      time.Time            `json:"modified,omitempty",bson:"modified"`
}
