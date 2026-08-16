package knife

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
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

	// Knife4j 前端（knife4j-vue）会基于 doc.html 所在路径推导前缀并自行拼接
	// （a + url），因此 swagger-config 中的 url/configUrl 必须是无前缀相对路径。
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dagine/v3/api-docs/swagger-config", nil))
	configBody := response.Body.String()
	if !strings.Contains(configBody, `"url": "/v3/api-docs"`) {
		t.Errorf("expected prefix-free url in swagger-config, got %q", configBody)
	}
	if strings.Contains(configBody, `"url": "/dagine`) {
		t.Errorf("expected no service prefix in swagger-config url, got %q", configBody)
	}
}

func TestInitHumaKnifeRegistersUIAndDocument(t *testing.T) {
	handler, api := humatest.New(t)
	if err := InitHumaKnife(api, Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	for _, testCase := range []struct {
		path        string
		wantContent []byte
		bodySnippet string
	}{
		{path: "/v3/api-docs", wantContent: []byte(openAPI303Document)},
		{path: "/doc.html", bodySnippet: "knife4j-vue"},
		{path: "/v3/api-docs/swagger-config", bodySnippet: `"url": "/v3/api-docs"`},
		{path: "/webjars/css/app.ac23e017.css", bodySnippet: "@charset"},
		{path: "/webjars/js/app.2fab4ac5.js", bodySnippet: ""},
		{path: "/favicon.ico", bodySnippet: ""},
	} {
		t.Run(testCase.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", response.Code)
			}
			if testCase.wantContent != nil && !bytes.Equal(response.Body.Bytes(), testCase.wantContent) {
				t.Fatalf("expected exact content %q, got %q", testCase.wantContent, response.Body.Bytes())
			}
			if testCase.bodySnippet != "" && !strings.Contains(response.Body.String(), testCase.bodySnippet) {
				t.Fatalf("expected response to contain %q, got %q", testCase.bodySnippet, response.Body.String())
			}
		})
	}
}

func TestStaticRoutesAreHiddenFromOpenAPI(t *testing.T) {
	_, api := humatest.New(t)
	if err := InitHumaKnife(api, Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}
	paths := api.OpenAPI().Paths
	for _, p := range []string{"/doc.html", "/v3/api-docs", "/favicon.ico"} {
		if _, ok := paths[p]; ok {
			t.Errorf("expected knife4go static route %q to be hidden from OpenAPI document", p)
		}
	}
}
