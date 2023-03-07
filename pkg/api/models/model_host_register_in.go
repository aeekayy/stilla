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

// HostRegisterIn ...
type HostRegisterIn struct {
	Name string   `form:"name" json:"name" yaml:"name"`
	Tags []string `form:"tags" json:"tags" yaml:"tags"`
}
