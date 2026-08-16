package huma

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	knife "github.com/jasonlabz/knife4go"
)

const openAPI303Document = `{"openapi":"3.0.3","info":{"title":"test","version":"1.0.0"},"paths":{}}`

func TestInitSwaggerKnifeRegistersUIAndDocument(t *testing.T) {
	handler, api := humatest.New(t)
	if err := InitSwaggerKnife(api, knife.Doc(openAPI303Document)); err != nil {
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
	if err := InitSwaggerKnife(api, knife.Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}
	paths := api.OpenAPI().Paths
	for _, p := range []string{"/doc.html", "/v3/api-docs", "/favicon.ico"} {
		if _, ok := paths[p]; ok {
			t.Errorf("expected knife4go static route %q to be hidden from OpenAPI document", p)
		}
	}
}
