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
	"strings"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"

	"github.com/aeekayy/stilla/service/pkg/utils"
)

const hostKey = "host"

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc func(*DAL) gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

// NewRouter returns a new router.
func NewRouter(dal *DAL) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies([]string{})

	// Setup the cookie store for session management
	// TODO: Make this optional
	store, err := sessions.NewRedisStore(10, "tcp", dal.Config.Cache.Host, dal.Config.Cache.Password, []byte(dal.SessionKey))

	if err != nil {
		dal.Logger.Errorf("error setting up cache for DAL: %s", err)
	} else {
		router.Use(sessions.Sessions("stilla", store))
	}

	if dal.Config.Sentry.Enabled {
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}

	if dal.Config.NewRelic.Enabled {
		router.Use(nrgin.Middleware(dal.APM))
	}

	authRequired := AuthRequired(dal)

	// Simple group: v1
	hostGroup := router.Group("/api/v1/host")
	hostGroup.Use(authRequired)
	for _, route := range hostRoutes {
		handler := route.HandlerFunc(dal)
		switch route.Method {
		case http.MethodGet:
			hostGroup.GET(route.Pattern, handler)
		case http.MethodPost:
			hostGroup.POST(route.Pattern, handler)
		case http.MethodPut:
			hostGroup.PUT(route.Pattern, handler)
		case http.MethodPatch:
			hostGroup.PATCH(route.Pattern, handler)
		case http.MethodDelete:
			hostGroup.DELETE(route.Pattern, handler)
		}
	}

	healthGroup := router.Group("/api/v1/health")
	for _, route := range healthRoutes {
		handler := route.HandlerFunc(dal)
		switch route.Method {
		case http.MethodGet:
			healthGroup.GET(route.Pattern, handler)
		case http.MethodPost:
			healthGroup.POST(route.Pattern, handler)
		case http.MethodPut:
			healthGroup.PUT(route.Pattern, handler)
		case http.MethodPatch:
			healthGroup.PATCH(route.Pattern, handler)
		case http.MethodDelete:
			healthGroup.DELETE(route.Pattern, handler)
		}
	}

	recordGroup := router.Group("/api/v1/records")
	recordGroup.Use(authRequired)
	for _, route := range recordRoutes {
		handler := route.HandlerFunc(dal)
		switch route.Method {
		case http.MethodGet:
			recordGroup.GET(route.Pattern, handler)
		case http.MethodPost:
			recordGroup.POST(route.Pattern, handler)
		case http.MethodPut:
			recordGroup.PUT(route.Pattern, handler)
		case http.MethodPatch:
			recordGroup.PATCH(route.Pattern, handler)
		case http.MethodDelete:
			recordGroup.DELETE(route.Pattern, handler)
		}
	}

	configGroup := router.Group("/api/v1/config")
	configGroup.Use(authRequired)
	for _, route := range configRoutes {
		handler := route.HandlerFunc(dal)
		switch route.Method {
		case http.MethodGet:
			configGroup.GET(route.Pattern, handler)
		case http.MethodPost:
			configGroup.POST(route.Pattern, handler)
		case http.MethodPut:
			configGroup.PUT(route.Pattern, handler)
		case http.MethodPatch:
			configGroup.PATCH(route.Pattern, handler)
		case http.MethodDelete:
			configGroup.DELETE(route.Pattern, handler)
		}
	}

	configsGroup := router.Group("/api/v1/configs")
	configsGroup.Use(authRequired)
	for _, route := range configsRoutes {
		handler := route.HandlerFunc(dal)
		switch route.Method {
		case http.MethodGet:
			configsGroup.GET(route.Pattern, handler)
		case http.MethodPost:
			configsGroup.POST(route.Pattern, handler)
		case http.MethodPut:
			configsGroup.PUT(route.Pattern, handler)
		case http.MethodPatch:
			configsGroup.PATCH(route.Pattern, handler)
		case http.MethodDelete:
			configsGroup.DELETE(route.Pattern, handler)
		}
	}

	return router
}

var hostRoutes = Routes{
	{
		"HostRegister",
		http.MethodPost,
		"/register",
		HostRegister,
	},

	{
		"HostLogin",
		http.MethodPost,
		"/login",
		HostLogin,
	},

	{
		"GetConfigByHostID",
		http.MethodGet,
		"/:hostId/config/:configId",
		GetConfigByID,
	},
}

var healthRoutes = Routes{
	{
		"PingGet",
		http.MethodGet,
		"/",
		PingGet,
	},
}

var recordRoutes = Routes{
	{
		"GetRecords",
		http.MethodGet,
		"/records",
		GetRecords,
	},
}

var configRoutes = Routes{
	{
		"AddConfig",
		http.MethodPost,
		"/",
		AddConfig,
	},

	{
		"GetConfigByID",
		http.MethodGet,
		"/:configId",
		GetConfigByID,
	},

	{
		"UpdateConfigByID",
		http.MethodPatch,
		"/:configId",
		UpdateConfigByID,
	},
}
var configsRoutes = Routes{
	{
		"GetConfigs",
		http.MethodGet,
		"/",
		GetConfigs,
	},
}

func extractToken(c *gin.Context) (string, string, bool) {
	bearerToken := c.Request.Header.Get("Authorization")
	host := c.Request.Header.Get("HostID")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1], host, true
	}

	return "", host, false
}

// AuthRequired is a simple middleware to check the session
func AuthRequired(d *DAL) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var host interface{}
		d.Logger.Info("Checking authorization")
		token, hostID, ok := extractToken(c)
		if ok {
			// check for the token's validity
			host, ok, _ = ValidateToken(d, hostID, token)
			c.Set("x-host-id", hostID)
			if !ok {
				d.Logger.Infof("Auth failed for %s", utils.ObfuscateValue(hostID, 8))
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
		} else {
			session := sessions.Default(c)
			host = session.Get(hostKey)
		}

		if host == "" {
			// Abort the request with the appropriate error code
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// set the context
		c.Set("x-host", host)

		c.Next()
	}

	return gin.HandlerFunc(fn)
}
