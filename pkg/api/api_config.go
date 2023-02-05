/*
 * Buffet Config Manager
 *
 * A configuration service that stores and retrieves configuration.
 *
 * API version: 0.1.0
 * Contact: apiteam@swagger.io
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

import (
	"context"
	"net/http"

	"github.com/aeekayy/stilla/pkg/api/models"
	"github.com/gin-gonic/gin"
)

// AddConfig - Create a new configuration and configuration value
func AddConfig(dal *DAL) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var req models.ConfigIn

		if err := c.ShouldBind(&req); err != nil {
			dal.Logger.Errorf("unable to parse request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to insert configuration"})
			return
		}

		config_id, err := dal.InsertConfig(context.TODO(), req, c.Request)

		if err != nil {
			dal.Logger.Errorf("unable to insert config: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to insert configuration"})
			return
		}

		dal.Logger.Infof("created config object %s", config_id)
		c.JSON(http.StatusCreated, gin.H{
			"data": config_id,
		})
	}

	return fn
}

// GetConfigByID - Retrieve a configuration by configuration ID
func GetConfigByID(dal *DAL) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		configID := c.Param("configId")
		hostID := c.Param("hostId")

		if configID == "" {
			dal.Logger.Errorf("unable to parse request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to retrieve configuration"})
			return
		}

		config, err := dal.GetConfig(context.TODO(), configID, hostID, c.Request)

		if err != nil {
			dal.Logger.Errorf("unable to retrieve config: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to retrieve configuration"})
			return
		}

		dal.Logger.Infof("retrieved config")
		c.JSON(http.StatusOK, gin.H{
			"data": config,
		})
	}

	return fn
}

// GetConfigs - Get a paginated list of configurations
func GetConfigs(dal *DAL) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		offset := c.Query("offset")
		limit := c.Query("limit")

		configs, err := dal.GetConfigs(context.TODO(), offset, limit, c.Request)

		if err != nil {
			dal.Logger.Errorf("unable to retrieve configurations: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to retrieve configurations"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": configs,
		})
	}

	return gin.HandlerFunc(fn)
}

// UpdateConfigByID - Update a configuration by configuration ID
func UpdateConfigByID(dal *DAL) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		configID := c.Param("configId")
		var req models.UpdateConfigIn

		if configID == "" {
			dal.Logger.Errorf("unable to parse request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to update configuration"})
			return
		}

		if err := c.ShouldBind(&req); err != nil {
			dal.Logger.Errorf("unable to parse request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to update configuration"})
			return
		}

		config, err := dal.UpdateConfigByID(context.TODO(), configID, req, c.Request)

		if err != nil {
			dal.Logger.Errorf("unable to retrieve config: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to retrieve configuration"})
			return
		}

		dal.Logger.Infof("retrieved config")
		c.JSON(http.StatusOK, gin.H{
			"data": config,
		})
	}

	return fn
}
