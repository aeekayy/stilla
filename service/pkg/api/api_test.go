// Package api the main package for the HTTP server
package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/aeekayy/stilla/service/pkg/models"
)

const (
	pageNotFoundErrMsg = "404 page not found"
	v1ApiPrefix = "/api/v1"
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

// TestAPIRoutes tests all of the API routes for the backend.
// This includes negative tests to make sure that we receive the 
// correct error codes. 
func TestAPIRoutes(t *testing.T) {
	dal := setupDep(t)
	router := NewRouter(dal)

	// setting up other dependencies such as request bodys
	rbHostRegister := strings.NewReader(`{ "name": "optimus-prime", "tags": ["autobot", "leader"] }`)

	table := []struct {
		name               string
		method             string
		path               string
		requestBody        io.Reader
		expectResponseCode int
		expectResponseBody string
	}{
		{"testPingRoutePositive", http.MethodGet, "/health/", nil, http.StatusOK, "{\"message\":\"pong\"}"},
		{"testBadPathNegative", http.MethodGet, "/bad-path", nil, http.StatusNotFound, pageNotFoundErrMsg },
		{"testRegisterHostPositive", http.MethodPost, "/host/register", rbHostRegister, http.StatusCreated, ""},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			fullPath := fmt.Sprintf("%s%s", v1ApiPrefix, tc.path)
			req, _ := http.NewRequest(tc.method, fullPath, tc.requestBody)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectResponseCode, w.Code)
			assert.Equal(t, tc.expectResponseBody, w.Body.String())
		})
	}
}
