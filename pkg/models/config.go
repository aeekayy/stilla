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
	Environment string                 `yaml:"enviornment", json:"environment"`
	Database    Database               `yaml:"database", json:"database"`
	Cache       Cache                  `yaml:"cache", json:"cache"`
	DocDB       DocumentDatabase       `yaml:"docdb", json:"docdb"`
	Server      Server                 `yaml:"server",json:"server"`
	Kafka       map[string]interface{} `yaml:"kafka", json:"kafka"`
	Audit       bool                   `yaml:"audit", json:"audit"`
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
	Host     string `yaml:"host", json:"host"`
	Username string `yaml:"username", json:"username"`
	Password string `yaml:"password", json:"password"`
	Type     string `yaml:"type", json:"type"`
}

// Cache struct to hold Postgres configuration
type Database struct {
	Username   string `yaml:"username",json:"username"`
	Password   string `yaml:"password",json:"password"`
	Host       string `yaml:"host",json:"host"`
	Name       string `yaml:"name",json:"name"`
	Parameters string `yaml:"parameters",json:"parameters"`
}

// Cache struct to hold MongoDB configuration
type DocumentDatabase struct {
	Username string `yaml:"username",json:"username"`
	Password string `yaml:"password",json:"password"`
	Host     string `yaml:"host",json:"host"`
	Name     string `yaml:"name",json:"name"`
	Timeout  string `yaml:"timeout",json:"timeout"`
}

// Server struct to hold web server configuration
type Server struct {
	Port    int    `yaml:"port",json:"port"`
	Timeout string `yaml:"timeout",json:"timeout"`
}

// GetConfig retrieves the Viper configuration for the service
func GetConfig(in string) (*Config, error) {
	// retrieve the configuration using viper
	viper.SetConfigName("stilla")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	if in != "" {
		viper.SetConfigFile(in)
	}
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
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
