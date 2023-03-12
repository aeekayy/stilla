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
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/aeekayy/stilla/pkg/api/models"
	"github.com/aeekayy/stilla/pkg/utils"
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

		_, apiKey, err := dal.RegisterHost(c, req, c.Request)

		if err != nil {
			output := utils.SanitizeLogMessage(req.Name, req.Name)
			dal.Logger.Errorf("unable to register host: %v", output)
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to register host"})
			return
		}

		// Todo replace the response with a struct. Include the host ID in the response
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

		hostID, err := dal.LoginHost(c, req, c.Request)

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
