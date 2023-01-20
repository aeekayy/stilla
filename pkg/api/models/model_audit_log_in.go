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

type AuditLogIn struct {
	Referer *interface{} `json:"referer,omitempty"`

	Service *interface{} `json:"service,omitempty"`

	Uri *interface{} `json:"uri,omitempty"`

	Host *interface{} `json:"host"`

	Operation *interface{} `json:"operation"`

	Auth *interface{} `json:"auth,omitempty"`

	Entity *interface{} `json:"entity,omitempty"`

	RequestType *interface{} `json:"request_type,omitempty"`

	Protocol *interface{} `json:"protocol,omitempty"`

	Headers *interface{} `json:"headers,omitempty"`
}
