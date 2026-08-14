package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAddOpenAPIDocumentRouterServesOpenAPIDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	const document = `{"openapi":"3.0.3","info":{"title":"test","version":"1.0.0"},"paths":{}}`
	AddOpenAPIDocumentRouter(router, "", document)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, API_DOCS_RELATIVE_PATH, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Body.String() != document {
		t.Fatalf("expected OpenAPI document %q, got %q", document, response.Body.String())
	}
}

func TestAddOpenAPIConfigRouterServesDocumentURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	AddOpenAPIConfigRouter(router, "", "/catalog/v3/api-docs")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, OPENAPI_CONFIG_PATH, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"url": "/catalog/v3/api-docs"`) {
		t.Fatalf("expected OpenAPI config to contain the document URL, got %q", response.Body.String())
	}
}
