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
	"github.com/gin-gonic/contrib/sessions"
)

// HostRegister - Register host for an API key
func HostRegister(dal *DAL) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var req models.HostRegisterIn

		if err := c.ShouldBind(&req); err != nil {
			dal.Logger.Errorf("unable to parse request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to register host"})
			return
		}

		apiKey, err := dal.RegisterHost(context.TODO(), req, c.Request)

		if err != nil {
			dal.Logger.Errorf("unable to register host: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to register host"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": apiKey,
		})
	}

	return gin.HandlerFunc(fn)
}

// HostLogin - Login host with an API key
func HostLogin(dal *DAL) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var req models.HostLoginIn

		if err := c.ShouldBind(&req); err != nil {
			dal.Logger.Errorf("unable to parse request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to login host"})
			return
		}

		hostID, err := dal.LoginHost(context.TODO(), req, c.Request)

		session := sessions.Default(c)

		if err != nil {
			dal.Logger.Errorf("unable to login host: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to login host"})
			return
		}

		// Save the host ID in the session
		session.Set(hostKey, hostID) // In real world usage you'd set this to the users ID
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": hostID,
		})
	}

	return gin.HandlerFunc(fn)
}
