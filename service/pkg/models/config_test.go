// Package models configuration model
package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
)

var (
	testKafkaConfig = map[string]string{
		"bootstrap.servers": "smitty:9092,lucky:9092",
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     "test",
		"sasl.password":     "kafka",
	}
)

// GetTestConfig gets a configuration for tests
func GetTestConfig() *Config {
	src := GetTestKafkaMap()
	testConfig := Config{
		Kafka: src,
		Audit: false,
		Server: Server{
			Port: 8080,
		},
		Database: Database{
			Username: "postgres",
			Password: "postgres",
			Host:     "localhost",
			Name:     "stilla",
		},
	}

	return &testConfig
}

// GetTestKafkaConfig gets the test Kafka configuration
func GetTestKafkaConfig() kafka.ConfigMap {
	cm := kafka.ConfigMap{}
	for k, v := range testKafkaConfig {
		cm.SetKey(k, v)
	}

	return cm
}

// GetTestKafkaMap gets the test Kakfa map[string]interface
func GetTestKafkaMap() map[string]interface{} {
	b := []byte(`{"bootstrap":{ "servers":"smitty:9092,lucky:9092" }, "security": {"protocol":"SASL_SSL"}, "sasl": {"mechanisms": "PLAIN", "username": "test", "password": "kafka" } }`)

	var i interface{}
	json.Unmarshal(b, &i)
	src := i.(map[string]interface{})
	return src
}

// TestNewConfig test the creation of a new configuration
func TestNewConfig(t *testing.T) {
	kafkaCfg := make(map[string]interface{})
	testConfig := Config{
		Kafka: kafkaCfg,
		Audit: false,
		Server: Server{
			Port: 8080,
		},
	}

	config := NewConfig()

	assert.Equal(t, *config, testConfig, "the configurations should match.")
}

// TestFlattenKafkaConfigMap test the flattening and conversion of a map to a Kafka ConfigMap
func TestFlattenKafkaConfigMap(t *testing.T) {
	src := GetTestKafkaMap()

	// create the config map
	cm := kafka.ConfigMap{}
	for k, v := range testKafkaConfig {
		cm.SetKey(k, v)
	}

	producedCm := kafka.ConfigMap{}

	flattenKafkaConfigMap("", src, &producedCm)

	assert.Equal(t, cm, producedCm, "the config maps should match.")
}

// TestGetKafkaConfig test GetKafkaConfig from Config
func TestGetKafkaConfig(t *testing.T) {
	config := GetTestConfig()
	kafkaCM := GetTestKafkaConfig()

	kafkaCfg := config.GetKafkaConfig()

	assert.Equal(t, kafkaCM, *kafkaCfg, "the Kafka config maps should match.")
}

// TestFailNoFileGetConfig test GetConfig for a file that doesn't exist
func TestFailNoFileGetConfig(t *testing.T) {
	fileIn := "tests/file-doesnt-exist.yaml"

	config, err := GetConfig(fileIn)

	// assert for nil
	assert.Nil(t, config)

	// assert equality for error
	checkErr := fmt.Errorf("open %s: no such file or directory", fileIn)
	assert.Equal(t, err, checkErr, "the errors should match.")
}

// TestPassGetConfig test GetConfig with a pass
func TestPassGetConfig(t *testing.T) {
	fileIn := "tests/stilla.pass.yaml"

	config, err := GetConfig(fileIn)

	assert.Nil(t, err)

	checkConfig := &Config{
		Kafka: map[string]interface{}{
			"bootstrap": map[string]interface{}{
				"servers": "localhost:9092",
			},
			"session": map[string]interface{}{
				"timeout": map[string]interface{}{
					"ms": 45000,
				},
			},
		},
		Database: Database{
			Username:   "postgres",
			Password:   "postgres",
			Host:       "localhost",
			Name:       "stilla",
			Parameters: "",
		},
		Cache: Cache{
			Host:     "localhost",
			Username: "default",
			Password: "",
			Type:     "redis",
		},
		Server: Server{
			Timeout: "3s",
			Port:    8080,
		},
		Sentry: SentryConfig{
			DSN:     "",
			Enabled: false,
		},
		Environment: "dev",
		SessionKey:  "",
		DocDB: DocumentDatabase{
			Username: "mongo",
			Password: "mongo",
			Host:     "localhost",
			Name:     "configdb",
			Timeout:  "",
			DNSSeed:  false,
		},
		NewRelic: NewRelicConfig{
			AppName:          "",
			License:          "",
			Enabled:          false,
			AppLogForwarding: false,
		},
		Audit: true,
	}

	assert.Equal(t, checkConfig, config, "the configurations should match.")
}
