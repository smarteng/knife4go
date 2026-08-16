package knife

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

const openAPI303Document = `{"openapi":"3.0.3","info":{"title":"test","version":"1.0.0"},"paths":{}}`

type fakeSwagger struct{ doc string }

func (f fakeSwagger) ReadDoc() string { return f.doc }

func TestInitSwaggerKnifeRegistersUIAndDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := InitSwaggerKnife(router, Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	for _, testCase := range []struct {
		path         string
		wantDocument []byte
		bodySnippet  string
	}{
		{path: "/v3/api-docs", wantDocument: []byte(openAPI303Document)},
		{path: "/doc.html", bodySnippet: "knife4j-vue"},
		{path: "/v3/api-docs/swagger-config", bodySnippet: `"url": "/v3/api-docs"`},
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
				t.Fatalf("expected response to contain %q, got %q", testCase.bodySnippet, response.Body.String())
			}
		})
	}
}

func TestInitSwaggerKnifeReadsDocFromSwagByDefault(t *testing.T) {
	if swag.GetSwagger(swag.Name) == nil {
		swag.Register(swag.Name, fakeSwagger{doc: openAPI303Document})
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := InitSwaggerKnife(router); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v3/api-docs", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if !bytes.Equal(response.Body.Bytes(), []byte(openAPI303Document)) {
		t.Fatalf("expected default OpenAPI document %q, got %q", openAPI303Document, response.Body.Bytes())
	}
}

func TestInitSwaggerKnifeCustomDocPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := InitSwaggerKnife(router, Doc(openAPI303Document), DocPath("/catalog-docs.html")); err != nil {
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

func TestInitSwaggerKnifeFailsWithoutDocument(t *testing.T) {
	if swag.GetSwagger(swag.Name) != nil {
		t.Skip("swagger document already registered by another test in this package")
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := InitSwaggerKnife(router); err == nil {
		t.Fatal("expected error when no document is available, got nil")
	}
}

func TestInitSwaggerKnifeUnderPrefixGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/dagine")
	if err := InitSwaggerKnife(group, Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	for _, path := range []string{
		"/dagine/doc.html",
		"/dagine/v3/api-docs",
		"/dagine/v3/api-docs/swagger-config",
		"/dagine/webjars/css/app.ac23e017.css",
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 for %s, got %d", path, response.Code)
		}
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dagine/v3/api-docs/swagger-config", nil))
	if !strings.Contains(response.Body.String(), `"url": "/dagine/v3/api-docs"`) {
		t.Errorf("expected prefixed url in swagger-config, got %q", response.Body.String())
	}
}
