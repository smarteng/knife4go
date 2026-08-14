package knife

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const openAPI303Document = `{"openapi":"3.0.3","info":{"title":"test","version":"1.0.0"},"paths":{}}`

func TestInitRegistersOpenAPIDocumentAndUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := Init(router, DocumentProviderFunc(func() ([]byte, error) {
		return []byte(openAPI303Document), nil
	})); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	for _, testCase := range []struct {
		path         string
		wantDocument []byte
		bodySnippet  string
	}{
		{path: "/v3/api-docs", wantDocument: []byte(openAPI303Document)},
		{path: "/doc.html", bodySnippet: "knife4j-vue"},
	} {
		t.Run(testCase.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))

			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", response.Code)
			}
			if testCase.wantDocument != nil && !bytes.Equal(response.Body.Bytes(), testCase.wantDocument) {
				t.Fatalf("expected exact OpenAPI document %q, got %q", testCase.wantDocument, response.Body.Bytes())
			}
			if testCase.bodySnippet != "" && !strings.Contains(response.Body.String(), testCase.bodySnippet) {
				t.Fatalf("expected UI response to contain %q, got %q", testCase.bodySnippet, response.Body.String())
			}
		})
	}
}

func TestInitRegistersUIAtCustomDocumentPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := Init(router, DocumentProviderFunc(func() ([]byte, error) {
		return []byte(openAPI303Document), nil
	}), WithDocPath("/catalog-docs.html")); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	for _, testCase := range []struct {
		path       string
		wantStatus int
	}{
		{path: "/catalog-docs.html", wantStatus: http.StatusOK},
		{path: "/doc.html", wantStatus: http.StatusNotFound},
	} {
		t.Run(testCase.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))

			if response.Code != testCase.wantStatus {
				t.Fatalf("expected status %d, got %d", testCase.wantStatus, response.Code)
			}
		})
	}
}

func TestInitRejectsInvalidDocumentProviders(t *testing.T) {
	tests := []struct {
		name     string
		provider DocumentProvider
	}{
		{name: "nil provider", provider: nil},
		{
			name: "provider error",
			provider: DocumentProviderFunc(func() ([]byte, error) {
				return nil, errors.New("document unavailable")
			}),
		},
		{
			name: "blank document",
			provider: DocumentProviderFunc(func() ([]byte, error) {
				return []byte{}, nil
			}),
		},
		{
			name: "invalid JSON",
			provider: DocumentProviderFunc(func() ([]byte, error) {
				return []byte("{"), nil
			}),
		},
		{
			name: "OpenAPI 3.1 document",
			provider: DocumentProviderFunc(func() ([]byte, error) {
				return []byte(`{"openapi":"3.1.0","info":{"title":"test","version":"1.0.0"},"paths":{}}`), nil
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			if err := Init(gin.New(), test.provider); err == nil {
				t.Fatal("expected initialization error")
			}
		})
	}
}

func TestInitUsesGroupPrefixInUIConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/catalog")
	if err := Init(group, DocumentProviderFunc(func() ([]byte, error) {
		return []byte(openAPI303Document), nil
	})); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/catalog/v3/api-docs/swagger-config", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var config map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &config); err != nil {
		t.Fatalf("decode UI configuration: %v", err)
	}
	if got, want := config["url"], "/catalog/v3/api-docs"; got != want {
		t.Fatalf("expected fully prefixed OpenAPI document URL %q, got %q", want, got)
	}
}
