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

// Checksum ...
type Checksum string

// ConfigVersion ...
type ConfigVersion struct {
	Config   map[string]interface{} `json:"config" bson:"config"`
	Checksum Checksum               `form:"checksum" json:"checksum,omitempty" bson:"checksum,omitempty"`
}
