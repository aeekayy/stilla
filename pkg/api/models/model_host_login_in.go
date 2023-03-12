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

// HostLoginIn ...
type HostLoginIn struct {
	APIKey string `form:"apikey" json:"apikey" yaml:"apikey"`
	Host   string `form:"host" json:"host" yaml:"host"`
}
