package knife

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/gin-gonic/gin"
)

const openAPI303Document = `{"openapi":"3.0.3","info":{"title":"test","version":"1.0.0"},"paths":{}}`

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
		{path: "/swagger/doc.json", wantDocument: []byte(openAPI303Document)},
		{path: "/doc.html", bodySnippet: "knife4j-vue"},
		{path: "/v3/api-docs/swagger-config", bodySnippet: `"urls": [{"name": "default","url": "/swagger/doc.json"`},
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

// TestInitSwaggerKnifeCustomAPIDocsPath 校验 APIDocsPath 选项能同时改动
// 文档 JSON 端点自身路径以及 swagger-config 响应体里的 url 字段。
func TestInitSwaggerKnifeCustomAPIDocsPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	const customPath = "/api/openapi.json"
	if err := InitSwaggerKnife(router,
		Doc(openAPI303Document),
		APIDocsPath(customPath),
	); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	// 自定义路径提供文档 JSON
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, customPath, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 on custom API docs path, got %d", response.Code)
	}
	if !bytes.Equal(response.Body.Bytes(), []byte(openAPI303Document)) {
		t.Fatalf("expected exact OpenAPI document at custom path, got %q", response.Body.Bytes())
	}

	// 默认路径不再注册
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on default API docs path when overridden, got %d", response.Code)
	}

	// swagger-config 端点路径保持不变，但内部 url 指向自定义路径
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v3/api-docs/swagger-config", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 on swagger-config endpoint, got %d", response.Code)
	}
	wantURL := `"url": "` + customPath + `"`
	if !strings.Contains(response.Body.String(), wantURL) {
		t.Errorf("expected swagger-config to reference custom API docs path %s, got %q", wantURL, response.Body.String())
	}
	wantLocation := `"location": "` + customPath + `"`
	if !strings.Contains(response.Body.String(), wantLocation) {
		t.Errorf("expected swagger-config to reference custom API docs location %s, got %q", wantLocation, response.Body.String())
	}
}

// TestInitSwaggerKnifeDocPathAutoPropagatesPrefix 覆盖用户把 UI 页面挂到子目录
// （典型：DocPath("/swagger/index.html")）时，knife4go 自动把 uiPrefix 应用到
// 静态资产、swagger-config、oauth2-redirect 与 API 文档端点。
func TestInitSwaggerKnifeDocPathAutoPropagatesPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := InitSwaggerKnife(router,
		Doc(openAPI303Document),
		DocPath("/swagger/index.html"),
		APIDocsPath("/swagger/doc.json"),
	); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	// 全部 knife4go 相关路由都必须能命中（uiPrefix = /swagger）。
	for _, tc := range []struct {
		path        string
		bodySnippet string
	}{
		{path: "/swagger/index.html", bodySnippet: "knife4j-vue"},
		{path: "/swagger/doc.json", bodySnippet: `"openapi":"3.0.3"`},
		{path: "/swagger/v3/api-docs/swagger-config", bodySnippet: `"validatorUrl"`},
		{path: "/swagger/webjars/css/app.ac23e017.css", bodySnippet: "@charset"},
		{path: "/swagger/webjars/js/app.2fab4ac5.js", bodySnippet: ""},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tc.path, response.Code)
			}
			if tc.bodySnippet != "" && !strings.Contains(response.Body.String(), tc.bodySnippet) {
				t.Errorf("expected response for %s to contain %q, got %q", tc.path, tc.bodySnippet, response.Body.String())
			}
		})
	}

	// 根路径不应再注册（避免与业务路由冲突）。
	for _, p := range []string{"/doc.html", "/webjars/css/app.ac23e017.css", "/v3/api-docs/swagger-config"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, p, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("expected 404 at root path %s when UI is under /swagger, got %d", p, response.Code)
		}
	}

	// swagger-config 响应体里的 url/configUrl/oauth2RedirectUrl 必须是 strip 掉
	// uiPrefix 后的相对路径——knife4j 前端会基于 index.html 目录自动拼前缀。
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/swagger/v3/api-docs/swagger-config", nil))
	body := response.Body.String()
	for _, want := range []string{
		`"configUrl": "/v3/api-docs/swagger-config"`,
		`"oauth2RedirectUrl": "/swagger-ui/oauth2-redirect.html"`,
		`"urls": [{"name": "default","url": "/doc.json","location": "/doc.json"}]`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected swagger-config body to contain %s, got %q", want, body)
		}
	}
	// 特别校验：url 字段不能重复包含 uiPrefix（否则前端拼接后会变成 /swagger/swagger/doc.json）。
	if strings.Contains(body, `"url": "/swagger/`) {
		t.Errorf("swagger-config url must be prefix-free for knife4j auto-prepending, got %q", body)
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
		"/dagine/swagger/doc.json",
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
	if !strings.Contains(configBody, `"url": "/swagger/doc.json"`) {
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
		{path: "/swagger/doc.json", wantContent: []byte(openAPI303Document)},
		{path: "/doc.html", bodySnippet: "knife4j-vue"},
		{path: "/v3/api-docs/swagger-config", bodySnippet: `"urls": [{"name": "default","url": "/swagger/doc.json"`},
		{path: "/webjars/css/app.ac23e017.css", bodySnippet: "@charset"},
		{path: "/webjars/js/app.2fab4ac5.js", bodySnippet: ""},
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
	for _, p := range []string{"/doc.html", "/swagger/doc.json", "/webjars/css/app.ac23e017.css"} {
		if _, ok := paths[p]; ok {
			t.Errorf("expected knife4go static route %q to be hidden from OpenAPI document", p)
		}
	}
}

func TestInitHumaKnifeWithPrefixGroupUsesDocumentUI(t *testing.T) {
	handler, api := humatest.New(t)
	group := huma.NewGroup(api, "/generate-example-project")
	// huma 组把前缀写入文档 paths；Document UI 前端不拼 doc.html 前缀，
	// 展示与调试请求直接使用文档自带的完整路径。
	const humaDoc = `{"openapi":"3.0.3","info":{"title":"t","version":"1.0.0"},"paths":{"/generate-example-project/health-check":{"get":{}}}}`
	if err := InitHumaKnife(group, Doc(humaDoc)); err != nil {
		t.Fatalf("unexpected initialization error: %v", err)
	}

	// doc.html 在前缀组下可访问
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/generate-example-project/doc.html", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	// 文档原样注册（不剥离前缀、不改写 paths）
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/generate-example-project/swagger/doc.json", nil))
	if !strings.Contains(response.Body.String(), "/generate-example-project/health-check") {
		t.Errorf("expected document paths untouched, got %q", response.Body.String())
	}

	// 资产注册在 /webjars 命名空间（单套 UI）
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/generate-example-project/webjars/js/app.2fab4ac5.js", nil))
	if response.Code != http.StatusOK {
		t.Errorf("expected asset status 200, got %d", response.Code)
	}
}

// captureGinDebugOutput 在 gin.DebugMode 下捕获注册期间的 [GIN-debug] 输出。
// setup 内允许调用者注册路由；返回捕获到的完整字节。
func captureGinDebugOutput(t *testing.T, setup func()) []byte {
	t.Helper()
	originalMode := gin.Mode()
	originalWriter := gin.DefaultWriter
	gin.SetMode(gin.DebugMode)
	buf := &bytes.Buffer{}
	gin.DefaultWriter = buf
	defer func() {
		gin.SetMode(originalMode)
		gin.DefaultWriter = originalWriter
	}()
	setup()
	return buf.Bytes()
}

func TestInitSwaggerKnifeSuppressesStaticRouteDebugLogsByDefault(t *testing.T) {
	var router *gin.Engine
	output := captureGinDebugOutput(t, func() {
		router = gin.New()
		if err := InitSwaggerKnife(router, Doc(openAPI303Document)); err != nil {
			t.Fatalf("unexpected initialization error: %v", err)
		}
	})
	// 40 条静态资源以及 knife4j 约定的 swagger-config 探测端点应被静默。
	for _, unwanted := range [][]byte{
		[]byte("/webjars/"),
		[]byte("/v3/api-docs/swagger-config"),
	} {
		if bytes.Contains(output, unwanted) {
			t.Errorf("expected route %q to be silenced from GIN-debug output, got:\n%s", unwanted, output)
		}
	}
	// 用户可定制的动态路由（UI 页面、文档 JSON 端点）应保留日志，
	// 方便用户确认 knife4go 把它们注册到了哪里。
	for _, want := range [][]byte{
		[]byte("/doc.html"),
		[]byte("/swagger/doc.json"),
	} {
		if !bytes.Contains(output, want) {
			t.Errorf("expected dynamic route %q to remain in GIN-debug output, got:\n%s", want, output)
		}
	}
}

func TestInitSwaggerKnifeVerboseKeepsDebugLogs(t *testing.T) {
	var router *gin.Engine
	output := captureGinDebugOutput(t, func() {
		router = gin.New()
		if err := InitSwaggerKnife(router, Doc(openAPI303Document), Verbose(true)); err != nil {
			t.Fatalf("unexpected initialization error: %v", err)
		}
	})
	if !bytes.Contains(output, []byte("/webjars/")) {
		t.Errorf("expected verbose mode to keep static webjars debug logs, got:\n%s", output)
	}
}

func TestInitSwaggerKnifeDoesNotAffectSubsequentBusinessRouteLogs(t *testing.T) {
	var router *gin.Engine
	output := captureGinDebugOutput(t, func() {
		router = gin.New()
		if err := InitSwaggerKnife(router, Doc(openAPI303Document)); err != nil {
			t.Fatalf("unexpected initialization error: %v", err)
		}
		// 用户在 knife 注册之后自行注册业务路由，日志应该正常输出。
		router.GET("/api/hello", func(c *gin.Context) {})
	})
	if !bytes.Contains(output, []byte("/api/hello")) {
		t.Errorf("expected business route /api/hello to still appear in GIN-debug output, got:\n%s", output)
	}
}
