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

type UpdateConfigIn struct {
	// Unique name for the configuration
	ConfigName string                 `json:"config_name"`
	Requester  string                 `json:"requester,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Parents    []string               `json:"parents,omitempty"`
}
