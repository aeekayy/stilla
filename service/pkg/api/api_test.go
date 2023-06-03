// Package api the main package for the HTTP server
package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
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