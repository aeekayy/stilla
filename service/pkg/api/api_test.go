// Package api the main package for the HTTP server
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/aeekayy/stilla/service/pkg/models"
)

// GetTestGinContext creates a Gin context for tests
func GetTestGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
	}

	return ctx
}

func getConfig() *models.Config {
	return models.NewConfig()
}

func TestPingRoute(t *testing.T) {
	dal := setupDep(t)
	router := NewRouter(dal)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"message\":\"pong\"}", w.Body.String())
}