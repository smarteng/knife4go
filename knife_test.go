package knife

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitSwaggerKnifeRegistersSingleDocumentUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	const document = `{"swagger":"2.0","info":{"title":"test","version":"1.0.0"},"paths":{}}`
	if err := InitSwaggerKnife(router, Doc(document)); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	for _, testCase := range []struct {
		path        string
		bodySnippet string
	}{
		{path: "/v3/api-docs", bodySnippet: `"swagger":"2.0"`},
		{path: "/v3/api-docs/swagger-config", bodySnippet: `"url": "/v3/api-docs"`},
		{path: "/doc.html", bodySnippet: "knife4j-vue"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))

		if response.Code != http.StatusOK {
			t.Fatalf("%s: expected status 200, got %d", testCase.path, response.Code)
		}
		if !strings.Contains(response.Body.String(), testCase.bodySnippet) {
			t.Fatalf("%s: expected response to contain %q", testCase.path, testCase.bodySnippet)
		}
	}
}
