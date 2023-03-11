// Package models contains models that are not stored in a data store
// such as resource configuration and environment details
package models

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/viper"
)

// Config main configuration struct for the service
type Config struct {
	Environment string                 `yaml:"enviornment" json:"environment" mapstructure:"environment"`
	Database    Database               `yaml:"database" json:"database" mapstructure:"database"`
	Cache       Cache                  `yaml:"cache" json:"cache" mapstructure:"cache"`
	DocDB       DocumentDatabase       `yaml:"docdb" json:"docdb" mapstructure:"docdb"`
	Server      Server                 `yaml:"server" json:"server" mapstructure:"server"`
	Kafka       map[string]interface{} `yaml:"kafka" json:"kafka" mapstructure:"kafka"`
	Audit       bool                   `yaml:"audit" json:"audit" mapstructure:"audit"`
	SessionKey  string                 `yaml:"session_key" json:"session_key" mapstructure:"session_key"`
	Sentry      SentryConfig           `yaml:"sentry" json:"sentry" mapstructure:"sentry"`
	NewRelic    NewRelicConfig         `yaml:"new_relic" json:"new_relic" mapstructure:"new_relic"`
}

// SentryConfig configuration for Sentry
type SentryConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	DSN     string `yaml:"dsn" json:"dsn" mapstructure:"dsn"`
}

// NewRelicConfig configuration for New Relic
type NewRelicConfig struct {
	Enabled          bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	AppName          string `yaml:"app_name" json:"app_name" mapstructure:"app_name"`
	License          string `yaml:"license" json:"license" mapstructure:"license"`
	AppLogForwarding bool   `yaml:"app_log_forwarding" json:"app_log_forwarding" mapstructure:"app_log_forwarding"`
}

// GetKafkaConfig retrieves a Kafka.ConfigMap compatible struct from
// our configuration. Viper supports nested configuration. However, we
// need a flatten struct for Kafka
func (c *Config) GetKafkaConfig() *kafka.ConfigMap {
	cm := &kafka.ConfigMap{}

	flattenKafkaConfigMap("", c.Kafka, cm)

	return cm
}

// flattenKafkaConfigMap converts a nested struct into a flatten config map
// TODO this is specifcally for Kafka.ConfigMap. Maybe open this up to other
// structs
func flattenKafkaConfigMap(prefix string, src map[string]interface{}, cm *kafka.ConfigMap) {
	if prefix != "" {
		prefix += "."
	}

	for k, v := range src {
		switch child := v.(type) {
		case map[string]interface{}:
			flattenKafkaConfigMap(prefix+k, child, cm)
		default:
			cm.SetKey(prefix+k, child)
		}
	}
}

// Cache struct to hold Redis configuration
type Cache struct {
	Host     string `yaml:"host" json:"host" mapstructure:"host"`
	Username string `yaml:"username" json:"username" mapstructure:"username"`
	Password string `yaml:"password" json:"password" mapstructure:"password"`
	Type     string `yaml:"type" json:"type" mapstructure:"type"`
}

// Database Cache struct to hold Postgres configuration
type Database struct {
	Username   string `yaml:"username" json:"username" mapstructure:"username"`
	Password   string `yaml:"password" json:"password" mapstructure:"password"`
	Host       string `yaml:"host" json:"host" mapstructure:"host"`
	Name       string `yaml:"name" json:"name" mapstructure:"name"`
	Parameters string `yaml:"parameters" json:"parameters" mapstructure:"parameters"`
}

// DocumentDatabase Cache struct to hold MongoDB configuration
type DocumentDatabase struct {
	Username string `yaml:"username" json:"username" mapstructure:"username"`
	Password string `yaml:"password" json:"password" mapstructure:"password"`
	Host     string `yaml:"host" json:"host" mapstructure:"host"`
	Name     string `yaml:"name" json:"name" mapstructure:"name"`
	Timeout  string `yaml:"timeout" json:"timeout" mapstructure:"timeout"`
	DNSSeed  bool   `yaml:"dns_seed" json:"dns_seed" mapstructure:"dns_seed"`
}

// Server struct to hold web server configuration
type Server struct {
	Port    int    `yaml:"port" json:"port" mapstructure:"port"`
	Timeout string `yaml:"timeout" json:"timeout" mapstructure:"timeout"`
}

// GetConfig retrieves the Viper configuration for the service
func GetConfig(in string) (*Config, error) {
	// retrieve the configuration using viper
	viper.SetConfigName("stilla")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/secrets")
	if in != "" {
		viper.SetConfigFile(in)
	}

	err := viper.ReadInConfig()

	switch t := err.(type) {
	case viper.ConfigFileNotFoundError:
		viper.SetConfigType("env")
		// viper environment variables
		viper.SetEnvPrefix("stilla")
		viper.AutomaticEnv()
	case error:
		return nil, fmt.Errorf("%v, %v", t, err)
	default:
	}

	conf := &Config{}
	err = viper.Unmarshal(conf)

	viper.SetDefault("audit", false)
	viper.SetDefault("docdb.timeout", "15s")

	if err != nil {
		return nil, fmt.Errorf("unable to decode into config struct, %v", err)
	}

	return conf, err
}
