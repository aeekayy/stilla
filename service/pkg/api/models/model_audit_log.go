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
)

// AuditLog ...
type AuditLog struct {
	Created  time.Time              `json:"created,omitempty"`
	Body     map[string]interface{} `json:"body,omitempty"`
	ID       string                 `json:"id,omitempty"`
	Service  string                 `json:"service,omitempty"`
	Funcname string                 `json:"funcname,omitempty"`
}
