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
	responseLookupPrefix = "lookup:"
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
		{"testRegisterHostPositive", http.MethodPost, "/host/register", rbHostRegister, http.StatusCreated, fmt.Sprintf("%s%s", responseLookupPrefix, "ApiKey")},
		{"testRegisterHostNegative", http.MethodPost, "/host/register", nil, http.StatusBadRequest, "{\"error\":\"unable to register host\"}"},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			fullPath := fmt.Sprintf("%s%s", v1ApiPrefix, tc.path)
			req, _ := http.NewRequest(tc.method, fullPath, tc.requestBody)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectResponseCode, w.Code)
			if strings.HasPrefix(tc.expectResponseBody, responseLookupPrefix) {
				mDB := dal.Database.(mockDB)
				value := mDB.Lookup["ApiKey"]
				expectedResponse := fmt.Sprintf("{\"data\":\"%s\"}", value)
				assert.Equal(t, expectedResponse, w.Body.String())
			} else {
				assert.Equal(t, tc.expectResponseBody, w.Body.String())
			}
		})
	}

	mDB := dal.Database.(mockDB)
	rbHostLogin := strings.NewReader(fmt.Sprintf(`{ "host": "%s", "apikey": "%s" }`, mDB.Lookup["Hostname"], mDB.Lookup["ApiKey"]))
	responseHostLogin := fmt.Sprintf(`{"data":"%s"}`, mDB.Lookup["HostID"])

	tableSet2 := []struct {
		name               string
		method             string
		path               string
		requestBody        io.Reader
		expectResponseCode int
		expectResponseBody string
	}{
		{"testLoginHostPositive", http.MethodPost, "/host/login", rbHostLogin, http.StatusOK, responseHostLogin },
		{"testLoginHostNegative", http.MethodPost, "/host/login", nil, http.StatusBadRequest, "{\"error\":\"unable to login host\"}" },
	}

	for _, tc := range tableSet2 {
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
