package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAddApiDocRouterServesSwaggerDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	const document = `{"swagger":"2.0","info":{"title":"test","version":"1.0.0"},"paths":{}}`
	AddApiDocRouter(router, "", document)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, API_DOCS_RELATIVE_PATH, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Body.String() != document {
		t.Fatalf("expected Swagger document %q, got %q", document, response.Body.String())
	}
}

func TestAddSwaggerResourcesRouterServesDefaultConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	AddSwaggerResourcesRouter(router, "")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, SWAGGER_RESOURCES_CONFIG_PATH, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"url": "/v3/api-docs"`) {
		t.Fatalf("expected Swagger config to contain the API document URL, got %q", response.Body.String())
	}
}
