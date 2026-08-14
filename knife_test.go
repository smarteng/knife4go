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

func TestInitResolvesProviderOnlyOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	calls := 0
	provider := DocumentProviderFunc(func() ([]byte, error) {
		calls++
		return []byte(openAPI303Document), nil
	})

	if err := Init(router, provider); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected provider to be called once during initialization, got %d calls", calls)
	}

	for requestNumber := 0; requestNumber < 3; requestNumber++ {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v3/api-docs", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("request %d: expected status 200, got %d", requestNumber+1, response.Code)
		}
		if !bytes.Equal(response.Body.Bytes(), []byte(openAPI303Document)) {
			t.Fatalf("request %d: expected exact OpenAPI document %q, got %q", requestNumber+1, openAPI303Document, response.Body.Bytes())
		}
	}
	if calls != 1 {
		t.Fatalf("expected document requests not to resolve provider again, got %d calls", calls)
	}
}

func TestInitRejectsInvalidDocumentProviders(t *testing.T) {
	tests := []struct {
		name         string
		provider     DocumentProvider
		wantContains string
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
		{
			name: "missing OpenAPI root field",
			provider: DocumentProviderFunc(func() ([]byte, error) {
				return []byte(`{"info":{"title":"test","version":"1.0.0"},"paths":{}}`), nil
			}),
			wantContains: "missing openapi version",
		},
		{
			name: "non-string OpenAPI root field",
			provider: DocumentProviderFunc(func() ([]byte, error) {
				return []byte(`{"openapi":3,"info":{"title":"test","version":"1.0.0"},"paths":{}}`), nil
			}),
			wantContains: "invalid openapi version",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			err := Init(gin.New(), test.provider)
			if err == nil {
				t.Fatal("expected initialization error")
			}
			if test.wantContains != "" && !strings.Contains(err.Error(), test.wantContains) {
				t.Fatalf("expected error to contain %q, got %q", test.wantContains, err)
			}
		})
	}
}

func TestInitWrapsProviderError(t *testing.T) {
	sentinel := errors.New("document unavailable")
	err := Init(gin.New(), DocumentProviderFunc(func() ([]byte, error) {
		return nil, sentinel
	}))

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected Init error to wrap provider error %q, got %v", sentinel, err)
	}
	if !strings.Contains(err.Error(), "resolve OpenAPI document") {
		t.Fatalf("expected contextual provider resolution error, got %q", err)
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
