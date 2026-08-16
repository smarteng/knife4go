# knife4go 重构实现计划：框架无关核心 + gin/huma 适配器

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 knife4go 回退到 `InitSwaggerKnife` 简单 API 形态（对应 dagine-dashboard/server/routers/router.go:45 引用的 6386e39 版本），核心逻辑框架无关，gin/huma 各一个薄适配器，去掉 internal 层。

**Architecture:** 核心（根包 `knife`）定义最小接口 `Router{ GET(path, contentType string, content []byte) }` 与注册流程 `Register`；约 150 个前端资产文件改为数据驱动（`ui.Asset` 声明 + 分类汇总切片）；gin 适配器在根包（`InitSwaggerKnife(gin.IRouter, ...Opts)`，与 6386e39 兼容），huma 适配器在 `huma/` 子包（`huma.API.Handle` + `Operation.Hidden`，不污染 huma 的 OpenAPI 文档）。

**Tech Stack:** Go 1.25、gin v1.10.0、swaggo/swag v1.16.3、danielgtaylor/huma/v2 v2.39.1（仅 huma/ 子包引入）。

## Global Constraints

- 模块路径 `github.com/jasonlabz/knife4go`；go.mod `go 1.25`（huma v2.39.1 要求 go >= 1.25）。
- 根包公开 API 与 6386e39 完全一致：`InitSwaggerKnife(router gin.IRouter, opts ...Opts) error`、`Opts`/`Doc(doc)`/`DocPath(path)`、`Config`；默认文档来源 `swag.ReadDoc("swagger")`。
- 路由路径保持：`/doc.html`、`/v3/api-docs`、`/v3/api-docs/swagger-config`、全部 webjars/icon 资产路径（以各文件现有 `RELATIVE_PATH` 常量为准，不得改动路径字符串）。
- Content-Type 保留旧行为（对应原 utils 函数）：GetHtml → `text/html;charset=UTF-8`，GetCss → `text/css;charset=UTF-8`，GetJs → `application/javascript;charset=UTF-8`，GetOther → 空串（不设置 Content-Type 头）。html/css 内容为原文，js/fonts/img/icons/gz/license 内容为 16 进制串需 `ui.MustHex` 解码。
- 删除 `log.Fatal`；不再手工设置 Content-Length / Connection 头（由框架与 net/http 处理）。
- 全部导出标识符（类型、函数、方法、变量）按 GoDoc 风格写注释，以标识符名称开头；包级注释覆盖根包、`ui`、`huma` 及全部资产子包。
- 每个任务验证命令：`go build ./... && go test ./... && go vet ./... && gofmt -l .`（gofmt 输出为空）。
- Windows 环境，Bash 工具可用（git mv 等 POSIX 命令走 Bash 工具）。

---

### Task 1: 核心骨架（ui 包 + 根包核心 + gin 入口 + 测试重写）

**Files:**
- Modify: `go.mod`（恢复 swag、go 1.25、huma）
- Create: `ui/assets.go`、`ui/DOC_HTML.go`、`register.go`、`gin.go`、`opts.go`、`register_test.go`
- Rewrite: `knife.go`、`knife_test.go`
- Delete: `utils/gin_utils.go`、`constant/constant.go`（及空目录）、`internal/ui/api_docs.go`、`internal/ui/swagger_resources.go`、`internal/ui/routes_test.go`、`internal/ui/knife/img/icons/hex_test.go` 暂不删（Task 5）

**Interfaces:**
- Produces: `ui.Asset`、`ui.MustHex`、`ui.DocHtml`、`ui.DocHtmlRelativePath`；根包 `Router`、`Register`、`Opts`、`Doc`、`DocPath`、`Config`、`InitSwaggerKnife`；`allAssets()` 先返回空切片（后续任务填充）。

- [ ] **Step 1: 更新 go.mod**

```bash
cd /f/baidu/aiib-go/knife4go && mkdir -p ui && git mv internal/ui/knife/DOC_HTML.go ui/DOC_HTML.go
```

编辑 `go.mod`：

```go
module github.com/jasonlabz/knife4go

go 1.25

require (
	github.com/danielgtaylor/huma/v2 v2.39.1
	github.com/gin-gonic/gin v1.10.0
	github.com/swaggo/swag v1.16.3
)
```

- [ ] **Step 2: 创建 `ui/assets.go`**

```go
// Package ui 提供 knife4go 内置的前端资源：Knife4j UI 页面、WebJars 静态文件与站点图标。
package ui

import (
	"encoding/hex"
	"fmt"
)

// Asset 描述一条静态 GET 路由：固定路径、Content-Type 与响应内容。
// ContentType 为空串表示不设置 Content-Type 响应头。
type Asset struct {
	Path        string
	ContentType string
	Content     []byte
}

// MustHex 将 16 进制字符串解码为字节；仅在包初始化声明资产时使用，非法输入直接 panic。
func MustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("knife4go: invalid hex asset content: %v", err))
	}
	return b
}
```

- [ ] **Step 3: 改写 `ui/DOC_HTML.go`**（从 internal/ui/knife/DOC_HTML.go 迁移而来）

保留原有 `DocHtmlHexContent` 大字符串原样（去掉 `constant.RootPath + ` 拼接），删除 gin/constant/utils import 与 `AddRouterOfDocHtml` 函数：

```go
package ui

const (
	// DocHtmlRelativePath 是 Knife4j UI 页面的默认路径。
	DocHtmlRelativePath = "/doc.html"
	// DocHtmlContent 是 Knife4j UI 页面 HTML 内容。
	DocHtmlContent = `<!DOCTYPE html><html lang=en>...` // 原 DocHtmlHexContent 内容原样
)

// DocHtml 是 Knife4j UI 页面资产。
var DocHtml = Asset{
	Path:        DocHtmlRelativePath,
	ContentType: "text/html;charset=UTF-8",
	Content:     []byte(DocHtmlContent),
}
```

- [ ] **Step 4: 创建 `opts.go`**

```go
package knife

// Config 是 InitSwaggerKnife / Register 的注册配置。
type Config struct {
	docJson string
	docPath string
}

// Opts 以函数选项模式配置 knife4go 注册行为。
type Opts func(*Config)

// Doc 指定 OpenAPI 文档内容（JSON 字符串）。
// 未指定时默认读取 swag.ReadDoc("swagger")，需要空白导入 _ "项目/docs/swagger"。
func Doc(doc string) Opts {
	return func(c *Config) {
		c.docJson = doc
	}
}

// DocPath 指定 UI 页面路径，默认 /doc.html。
func DocPath(path string) Opts {
	return func(c *Config) {
		c.docPath = path
	}
}
```

- [ ] **Step 5: 创建 `register.go`**（框架无关核心）

```go
package knife

import (
	"fmt"
	"strings"

	"github.com/swaggo/swag"

	"github.com/jasonlabz/knife4go/ui"
)

// Router 是 knife4go 注册静态 GET 路由所需的最小框架无关接口。
// 新增 web/api 框架支持：实现该接口后调用 Register，或参照 huma/ 提供便捷入口包。
type Router interface {
	// GET 注册一条 GET 路由，响应体固定为 content，Content-Type 为 contentType（空串表示不设置）。
	GET(path, contentType string, content []byte)
}

const (
	// apiDocsPath 是 OpenAPI 文档的相对路径。
	apiDocsPath = "/v3/api-docs"
	// swaggerConfigSuffix 是 UI 配置端点的相对路径后缀。
	swaggerConfigSuffix = "/swagger-config"
	// oauth2RedirectPath 是 oauth2 重定向页面的相对路径。
	oauth2RedirectPath = "/swagger-ui/oauth2-redirect.html"
	// jsonContentType 是 JSON 响应使用的 Content-Type。
	jsonContentType = "application/json;charset=UTF-8"
)

// Register 将 knife4go 的 UI 页面、OpenAPI 文档端点与全部静态资产注册到 router。
func Register(router Router, opts ...Opts) error {
	config := Config{}
	for _, opt := range opts {
		opt(&config)
	}
	if config.docJson == "" {
		doc, err := swag.ReadDoc("swagger")
		if err != nil {
			return fmt.Errorf("read OpenAPI document from swag: %w", err)
		}
		config.docJson = doc
	}

	docURL := apiDocsPath
	if routerWithBasePath, ok := router.(interface{ BasePath() string }); ok {
		docURL = strings.TrimSuffix(routerWithBasePath.BasePath(), "/") + apiDocsPath
	}
	basePath := strings.TrimSuffix(docURL, apiDocsPath)

	// UI 页面（路径可经 DocPath 定制）
	docPath := config.docPath
	if docPath == "" {
		docPath = ui.DocHtmlRelativePath
	}
	router.GET(docPath, ui.DocHtml.ContentType, ui.DocHtml.Content)

	// OpenAPI 文档端点
	router.GET(docURL, jsonContentType, []byte(config.docJson))

	// UI 配置端点（Knife4j 前端据此加载文档与 oauth2 跳转）
	configContent := fmt.Sprintf(`{"configUrl": %q,"oauth2RedirectUrl": %q,"url": %q,"validatorUrl": ""}`,
		basePath+apiDocsPath+swaggerConfigSuffix,
		basePath+oauth2RedirectPath,
		docURL,
	)
	router.GET(basePath+apiDocsPath+swaggerConfigSuffix, jsonContentType, []byte(configContent))

	// 静态资产
	for _, asset := range allAssets() {
		router.GET(asset.Path, asset.ContentType, asset.Content)
	}
	return nil
}

// allAssets 返回全部静态资产；由各资产子包任务逐步填充。
func allAssets() []ui.Asset {
	return nil
}
```

- [ ] **Step 6: 创建 `gin.go`**（gin 适配器，根包）

```go
package knife

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginRouter 将 gin.IRouter 适配为 Router。
type ginRouter struct {
	r gin.IRouter
}

// GET 实现 Router.GET。
func (w *ginRouter) GET(path, contentType string, content []byte) {
	w.r.GET(path, func(c *gin.Context) {
		c.Data(http.StatusOK, contentType, content)
	})
}

// BasePath 返回路由组前缀，Register 据此计算 OpenAPI 文档的完整 URL。
func (w *ginRouter) BasePath() string {
	return w.r.BasePath()
}
```

- [ ] **Step 7: 重写 `knife.go`**（兼容入口，与 6386e39 签名一致）

```go
package knife

import "github.com/gin-gonic/gin"

// InitSwaggerKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 gin 路由组。
// 文档来源默认 swag.ReadDoc("swagger")（需空白导入 _ "项目/docs/swagger"），
// 也可通过 Doc(doc) 显式指定；UI 页面路径默认 /doc.html，可用 DocPath(path) 修改。
func InitSwaggerKnife(router gin.IRouter, opts ...Opts) error {
	return Register(&ginRouter{r: router}, opts...)
}
```

- [ ] **Step 8: 删除旧核心文件与旧资产源**

`internal/ui/knife` 下的旧资产文件仍 import 已删除的 `constant`/`utils`，必须整体移除（内容保留在 git 历史中，后续任务用 `git show ca6aa81:<path>` 取回转换（ca6aa81 是重构前最后一个代码提交，固定引用不受后续提交影响））：

```bash
cd /f/baidu/aiib-go/knife4go && git rm -q -r utils constant internal && rmdir utils constant internal 2>/dev/null; true
```

- [ ] **Step 9: 重写 `knife_test.go`**（gin 端到端）

```go
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
```

- [ ] **Step 10: 创建 `register_test.go`**（假 Router 回归核心）

```go
package knife

import (
	"bytes"
	"strings"
	"testing"
)

type fakeRoute struct {
	path        string
	contentType string
	content     []byte
}

type fakeRouter struct {
	routes []fakeRoute
}

func (f *fakeRouter) GET(path, contentType string, content []byte) {
	f.routes = append(f.routes, fakeRoute{path: path, contentType: contentType, content: content})
}

func fakeRouterRoutes(t *testing.T, opts ...Opts) []fakeRoute {
	t.Helper()
	router := &fakeRouter{}
	if err := Register(router, append([]Opts{Doc(openAPI303Document)}, opts...)...); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	return router.routes
}

func TestRegisterRegistersDynamicRoutes(t *testing.T) {
	routes := fakeRouterRoutes(t)

	for _, want := range []string{"/doc.html", "/v3/api-docs", "/v3/api-docs/swagger-config"} {
		if !containsRoute(routes, want) {
			t.Errorf("expected route %q to be registered", want)
		}
	}
	if got := routeContent(routes, "/v3/api-docs"); string(got) != openAPI303Document {
		t.Errorf("expected /v3/api-docs to serve the OpenAPI document, got %q", got)
	}
	if got := routeContent(routes, "/v3/api-docs/swagger-config"); !strings.Contains(string(got), `"url": "/v3/api-docs"`) {
		t.Errorf("expected swagger-config to contain the document URL, got %q", got)
	}
	if got := routeContent(routes, "/doc.html"); !bytes.Contains(got, []byte("knife4j-vue")) {
		t.Errorf("expected /doc.html to serve the UI page, got %q", got)
	}
}

func TestRegisterBasePathAffectsDocumentURL(t *testing.T) {
	router := &fakeRouter{}
	if err := Register(router, Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	// fakeRouter 未实现 BasePath，documentURL 保持根路径
	if got := routeContent(router.routes, "/v3/api-docs/swagger-config"); !strings.Contains(string(got), `"url": "/v3/api-docs"`) {
		t.Errorf("expected root document URL, got %q", got)
	}
}

type prefixedFakeRouter struct {
	fakeRouter
	basePath string
}

func (p *prefixedFakeRouter) BasePath() string { return p.basePath }

func TestRegisterHonorsRouterBasePath(t *testing.T) {
	router := &prefixedFakeRouter{basePath: "/catalog"}
	if err := Register(router, Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	if got := routeContent(router.routes, "/catalog/v3/api-docs/swagger-config"); !strings.Contains(string(got), `"url": "/catalog/v3/api-docs"`) {
		t.Errorf("expected prefixed document URL in config, got %q", got)
	}
	if got := routeContent(router.routes, "/catalog/v3/api-docs"); string(got) != openAPI303Document {
		t.Errorf("expected prefixed document route to serve the OpenAPI document, got %q", got)
	}
}

func containsRoute(routes []fakeRoute, path string) bool {
	for _, r := range routes {
		if r.path == path {
			return true
		}
	}
	return false
}

func routeContent(routes []fakeRoute, path string) []byte {
	for _, r := range routes {
		if r.path == path {
			return r.content
		}
	}
	return nil
}
```

- [ ] **Step 11: 验证与提交**

```bash
cd /f/baidu/aiib-go/knife4go && go mod tidy && go build ./... && go test ./... && go vet ./... && gofmt -l .
```

Expected: 全部通过，gofmt 无输出。

提交：

```bash
git add -A && git commit -m "feat: framework-agnostic core with gin adapter, restore InitSwaggerKnife API"
```

---

### Task 2: 迁移 webjars/css（5 个资产）

**Files:**
- Create: `ui/webjars/css/`（5 个资产文件 + `assets.go` 汇总）
- Modify: `register.go`（allAssets 加入 css 分类）、`register_test.go`（期望路径）

**Interfaces:**
- Consumes: `ui.Asset`、`ui.MustHex`
- Produces: `css.Assets []ui.Asset`

- [ ] **Step 1: 取回源文件并转换**

```bash
cd /f/baidu/aiib-go/knife4go && mkdir -p ui/webjars/css && for f in APP_AC23E017_CSS CHUNK_75464E7E_8FB93BA5_CSS CHUNK_D7D5F59C_A9FFBFCB_CSS CHUNK_VENDORS_F24A310A_CSS CHUNK_VENDORS_F24A310A_CSS_GZ; do git show ca6aa81:internal/ui/knife/webjars/css/$f.go > ui/webjars/css/$f.go; done
```

对每个文件执行相同转换（以 APP_AC23E017_CSS.go 为例，内容常量 `APP_AC23E017_CSS_HEX_CONTENT` 原样保留）：

```go
package css

import "github.com/jasonlabz/knife4go/ui"

const (
	// APP_AC23E017_CSS_RELATIVE_PATH 是 app.ac23e017.css 的路径。
	APP_AC23E017_CSS_RELATIVE_PATH = "/webjars/css/app.ac23e017.css"
	// APP_AC23E017_CSS_HEX_CONTENT 是 app.ac23e017.css 的内容。
	APP_AC23E017_CSS_HEX_CONTENT = `@charset "UTF-8";...` // 原内容原样
)

// AppAc23e017Css 是 app.ac23e017.css 资产。
var AppAc23e017Css = ui.Asset{
	Path:        APP_AC23E017_CSS_RELATIVE_PATH,
	ContentType: "text/css;charset=UTF-8",
	Content:     []byte(APP_AC23E017_CSS_HEX_CONTENT),
}
```

规则：删除 gin/constant/utils import 与 `AddRouterOfXxx` 函数；`RELATIVE_PATH` 去掉 `constant.RootPath + ` 前缀；`ContentType` 取原文件 utils 调用对应值（css → `text/css;charset=UTF-8`）；css 内容为原文直接 `[]byte(...)`；唯一例外 `CHUNK_VENDORS_F24A310A_CSS_GZ.go` 内容是 16 进制、原调用 GetOther，故 `ContentType: ""`、`Content: ui.MustHex(...)`。

- [ ] **Step 2: 创建 `ui/webjars/css/assets.go`**

```go
// Package css 提供 knife4go UI 使用的 css 静态资源。
package css

import "github.com/jasonlabz/knife4go/ui"

// Assets 是 css 分类的全部静态资源。
var Assets = []ui.Asset{
	AppAc23e017Css,
	Chunk75464e7e8fb93ba5Css,
	ChunkD7d5f59cA9ffbfcbCss,
	ChunkVendorsF24a310aCss,
	ChunkVendorsF24a310aCssGz,
}
```

- [ ] **Step 3: register.go 的 allAssets 加入 css**

```go
import (
	// ...现有 import
	"github.com/jasonlabz/knife4go/ui/webjars/css"
)

// allAssets 返回全部静态资产。
func allAssets() []ui.Asset {
	return append([]ui.Asset{}, css.Assets...)
}
```

- [ ] **Step 4: register_test.go 增加静态资产路径断言**

在 `TestRegisterRegistersDynamicRoutes` 后新增：

```go
func TestRegisterRegistersAllStaticAssets(t *testing.T) {
	routes := fakeRouterRoutes(t)
	// 期望路径以各资产文件 RELATIVE_PATH 常量为准（此处与常量一致）
	for _, want := range []string{
		"/webjars/css/app.ac23e017.css",
		"/webjars/css/chunk-75464e7e.8fb93ba5.css",
		"/webjars/css/chunk-d7d5f59c.a9ffbfcb.css",
		"/webjars/css/chunk-vendors.f24a310a.css",
		"/webjars/css/chunk-vendors.f24a310a.css.gz",
	} {
		if !containsRoute(routes, want) {
			t.Errorf("expected asset route %q to be registered", want)
		}
	}
}
```

- [ ] **Step 5: 验证与提交**

```bash
cd /f/baidu/aiib-go/knife4go && go build ./... && go test ./... && go vet ./... && gofmt -l . && git add -A && git commit -m "feat: data-driven webjars css assets"
```

---

### Task 3: 迁移 webjars/js（22 个资产）

**Files:**
- Create: `ui/webjars/js/`（22 个资产文件 + `assets.go` 汇总）
- Modify: `register.go`、`register_test.go`

**Interfaces:**
- Consumes: `ui.Asset`、`ui.MustHex`
- Produces: `js.Assets []ui.Asset`

- [ ] **Step 1: 取回源文件并转换**

```bash
cd /f/baidu/aiib-go/knife4go && mkdir -p ui/webjars/js && for f in APP_2FAB4AC5_JS CHUNK_069EB437_2CFEBF27_JS CHUNK_069EB437_2CFEBF27_JS_LICENSE_TXT CHUNK_0D102D5A_B2BDDFFC_JS CHUNK_0FD67716_D57E2C41_JS CHUNK_260D712A_390177FE_JS CHUNK_260D712A_390177FE_JS_LICENSE_TXT CHUNK_2D0AF44E_392AFCD6_JS CHUNK_2D0BD799_EB48B7F1_JS CHUNK_2D0DA532_591AD7FC_JS CHUNK_3B888A65_8737CE4F_JS CHUNK_3EC4AAA8_A79D19F8_JS CHUNK_589FAEE0_5BFD1708_JS CHUNK_589FAEE0_5BFD1708_JS_LICENSE_TXT CHUNK_735C675C_5B409314_JS CHUNK_75464E7E_B130271B_JS CHUNK_ADB9E944_2C7F24FE_JS CHUNK_ADB9E944_2C7F24FE_JS_LICENSE_TXT CHUNK_D7D5F59C_E61130F3_JS CHUNK_D7D5F59C_E61130F3_JS_LICENSE_TXT CHUNK_VENDORS_D51CF6F8_JS CHUNK_VENDORS_D51CF6F8_JS_LICENSE_TXT; do git show ca6aa81:internal/ui/knife/webjars/js/$f.go > ui/webjars/js/$f.go; done
```

转换模式同 Task 2，差异：
- 内容均为 16 进制 → `Content: ui.MustHex(XXX_HEX_CONTENT)`。
- `.js` 文件原调用 GetJs → `ContentType: "application/javascript;charset=UTF-8"`；`.js.LICENSE.txt` 文件原调用 GetOther → `ContentType: ""`。
- 资产变量名：去掉 `AddRouterOf` 前缀（如 `AddRouterOfApp2fab4ac5Js` → `App2fab4ac5Js`）。

- [ ] **Step 2: 创建 `ui/webjars/js/assets.go`**

```go
// Package js 提供 knife4go UI 使用的 js 静态资源。
package js

import "github.com/jasonlabz/knife4go/ui"

// Assets 是 js 分类的全部静态资源。
var Assets = []ui.Asset{
	App2fab4ac5Js,
	Chunk069eb4372cfebf27Js,
	Chunk069eb4372cfebf27JsLICENSETxt,
	Chunk0d102d5aB2bddffcJs,
	Chunk0fd67716D57e2c41Js,
	Chunk260d712a390177feJs,
	Chunk260d712a390177feJsLICENSETxt,
	Chunk2d0af44e392afcd6Js,
	Chunk2d0bd799Eb48b7f1Js,
	Chunk2d0da532591ad7fcJs,
	Chunk3b888a658737ce4fJs,
	Chunk3ec4aaa8A79d19f8Js,
	Chunk589faee05bfd1708Js,
	Chunk589faee05bfd1708JsLICENSETxt,
	Chunk735c675c5b409314Js,
	Chunk75464e7eB130271bJs,
	ChunkAdb9e9442c7f24feJs,
	ChunkAdb9e9442c7f24feJsLICENSETxt,
	ChunkD7d5f59cE61130f3Js,
	ChunkD7d5f59cE61130f3JsLICENSETxt,
	ChunkVendorsD51cf6f8Js,
	ChunkVendorsD51cf6f8JsLICENSETxt,
}
```

- [ ] **Step 3: register.go 的 allAssets 加入 js 分类**（import `"github.com/jasonlabz/knife4go/ui/webjars/js"`）

```go
// allAssets 返回全部静态资产。
func allAssets() []ui.Asset {
	assets := append([]ui.Asset{}, css.Assets...)
	return append(assets, js.Assets...)
}
```

- [ ] **Step 4: register_test.go 扩展 TestRegisterRegistersAllStaticAssets 期望路径**（追加 js 分类，路径以各文件 RELATIVE_PATH 常量为准）：

```go
		"/webjars/js/app.2fab4ac5.js",
		"/webjars/js/chunk-069eb437.2cfebf27.js",
		"/webjars/js/chunk-069eb437.2cfebf27.js.LICENSE.txt",
		"/webjars/js/chunk-0d102d5a.b2bddffc.js",
		"/webjars/js/chunk-0fd67716.d57e2c41.js",
		"/webjars/js/chunk-260d712a.390177fe.js",
		"/webjars/js/chunk-260d712a.390177fe.js.LICENSE.txt",
		"/webjars/js/chunk-2d0af44e.392afcd6.js",
		"/webjars/js/chunk-2d0bd799.eb48b7f1.js",
		"/webjars/js/chunk-2d0da532.591ad7fc.js",
		"/webjars/js/chunk-3b888a65.8737ce4f.js",
		"/webjars/js/chunk-3ec4aaa8.a79d19f8.js",
		"/webjars/js/chunk-589faee0.5bfd1708.js",
		"/webjars/js/chunk-589faee0.5bfd1708.js.LICENSE.txt",
		"/webjars/js/chunk-735c675c.5b409314.js",
		"/webjars/js/chunk-75464e7e.b130271b.js",
		"/webjars/js/chunk-adb9e944.2c7f24fe.js",
		"/webjars/js/chunk-adb9e944.2c7f24fe.js.LICENSE.txt",
		"/webjars/js/chunk-d7d5f59c.e61130f3.js",
		"/webjars/js/chunk-d7d5f59c.e61130f3.js.LICENSE.txt",
		"/webjars/js/chunk-vendors.d51cf6f8.js",
		"/webjars/js/chunk-vendors.d51cf6f8.js.LICENSE.txt",
```

- [ ] **Step 5: 验证与提交**

```bash
cd /f/baidu/aiib-go/knife4go && go build ./... && go test ./... && go vet ./... && gofmt -l . && git add -A && git commit -m "feat: data-driven webjars js assets"
```

---

### Task 4: 迁移 webjars/fonts（6）与 webjars/img（6）

**Files:**
- Create: `ui/webjars/fonts/`（6 个 + `assets.go`）、`ui/webjars/img/`（6 个 + `assets.go`）
- Modify: `register.go`、`register_test.go`

**Interfaces:**
- Consumes: `ui.Asset`、`ui.MustHex`
- Produces: `fonts.Assets []ui.Asset`、`img.Assets []ui.Asset`

- [ ] **Step 1: 取回并转换 fonts**

```bash
cd /f/baidu/aiib-go/knife4go && mkdir -p ui/webjars/fonts && for f in FONTAWESOME_WEBFONT_706450D7_TTF FONTAWESOME_WEBFONT_97493D3F_WOFF2 FONTAWESOME_WEBFONT_D9EE23D5_WOFF FONTAWESOME_WEBFONT_F7C2B4B7_EOT ICONFONT_4CA3D0C0_TTF ICONFONT_E2D2B98E_EOT; do git show ca6aa81:internal/ui/knife/webjars/fonts/$f.go > ui/webjars/fonts/$f.go; done
```

转换：内容 16 进制 → `ui.MustHex`；原调用 GetOther → `ContentType: ""`。资产变量名 `AddRouterOfFontawesomeWebfont706450d7Ttf` → `FontawesomeWebfont706450d7Ttf`，其余类推。

- [ ] **Step 2: 创建 `ui/webjars/fonts/assets.go`**

```go
// Package fonts 提供 knife4go UI 使用的字体资源。
package fonts

import "github.com/jasonlabz/knife4go/ui"

// Assets 是 fonts 分类的全部静态资源。
var Assets = []ui.Asset{
	FontawesomeWebfont706450d7Ttf,
	FontawesomeWebfont97493d3fWoff2,
	FontawesomeWebfontD9ee23d5Woff,
	FontawesomeWebfontF7c2b4b7Eot,
	Iconfont4ca3d0c0Ttf,
	IconfontE2d2b98eEot,
}
```

- [ ] **Step 3: 取回并转换 img**

```bash
cd /f/baidu/aiib-go/knife4go && mkdir -p ui/webjars/img && for f in EDITORMD_LOGO_53EA80E2_SVG FONTAWESOME_WEBFONT_29800836_SVG ICONFONT_1D48C203_SVG LOADING2X_695405A9_GIF LOADING3X_65EACF61_GIF LOADING_C929501E_GIF; do git show ca6aa81:internal/ui/knife/webjars/img/$f.go > ui/webjars/img/$f.go; done
```

转换同 fonts：`ui.MustHex`、`ContentType: ""`。

- [ ] **Step 4: 创建 `ui/webjars/img/assets.go`**

```go
// Package img 提供 knife4go UI 使用的图片资源。
package img

import "github.com/jasonlabz/knife4go/ui"

// Assets 是 img 分类的全部静态资源。
var Assets = []ui.Asset{
	EditormdLogo53ea80e2Svg,
	FontawesomeWebfont29800836Svg,
	Iconfont1d48c203Svg,
	Loading2x695405a9Gif,
	Loading3x65eacf61Gif,
	LoadingC929501eGif,
}
```

- [ ] **Step 5: register.go 加入 fonts 与 img**

```go
import (
	// ...现有 import
	"github.com/jasonlabz/knife4go/ui/webjars/fonts"
	"github.com/jasonlabz/knife4go/ui/webjars/img"
)

// allAssets 返回全部静态资产。
func allAssets() []ui.Asset {
	assets := append([]ui.Asset{}, css.Assets...)
	assets = append(assets, js.Assets...)
	assets = append(assets, fonts.Assets...)
	return append(assets, img.Assets...)
}
```

- [ ] **Step 6: register_test.go 追加 fonts 与 img 期望路径**

```go
		"/webjars/fonts/fontawesome-webfont.706450d7.ttf",
		"/webjars/fonts/fontawesome-webfont.97493d3f.woff2",
		"/webjars/fonts/fontawesome-webfont.d9ee23d5.woff",
		"/webjars/fonts/fontawesome-webfont.f7c2b4b7.eot",
		"/webjars/fonts/iconfont.4ca3d0c0.ttf",
		"/webjars/fonts/iconfont.e2d2b98e.eot",
		"/webjars/img/editormd-logo.53ea80e2.svg",
		"/webjars/img/fontawesome-webfont.29800836.svg",
		"/webjars/img/iconfont.1d48c203.svg",
		"/webjars/img/loading-2x.695405a9.gif",
		"/webjars/img/loading-3x.65eacf61.gif",
		"/webjars/img/loading.c929501e.gif",
```

（路径以各文件 RELATIVE_PATH 常量为准。）

- [ ] **Step 7: 验证与提交**

```bash
cd /f/baidu/aiib-go/knife4go && go build ./... && go test ./... && go vet ./... && gofmt -l . && git add -A && git commit -m "feat: data-driven webjars fonts and img assets"
```

---

### Task 5: 迁移 webjars/oauth（2）与 icons（14），清空 internal/

**Files:**
- Create: `ui/webjars/oauth/`（2 个 + `assets.go`）、`ui/icons/`（14 个 + `assets.go`）
- Delete: 原 `internal/ui/knife/img/icons/file.js`、`hex_test.go`（生成调试残留，不迁移）
- Modify: `register.go`、`register_test.go`

**Interfaces:**
- Consumes: `ui.Asset`、`ui.MustHex`
- Produces: `oauth.Assets []ui.Asset`、`icons.Assets []ui.Asset`

- [ ] **Step 1: 取回并转换 oauth**

```bash
cd /f/baidu/aiib-go/knife4go && mkdir -p ui/webjars/oauth && for f in AXIOS_MIN_JS OAUTH2_HTML; do git show ca6aa81:internal/ui/knife/webjars/oauth/$f.go > ui/webjars/oauth/$f.go; done
```

转换：`AXIOS_MIN_JS.go` 内容 16 进制 → `ui.MustHex`，原调用 GetJs → `ContentType: "application/javascript;charset=UTF-8"`；`OAUTH2_HTML.go` 内容原文 → `[]byte(...)`，原调用 GetHtml → `ContentType: "text/html;charset=UTF-8"`。

- [ ] **Step 2: 创建 `ui/webjars/oauth/assets.go`**

```go
// Package oauth 提供 knife4go UI 的 oauth2 登录辅助资源。
package oauth

import "github.com/jasonlabz/knife4go/ui"

// Assets 是 oauth 分类的全部静态资源。
var Assets = []ui.Asset{
	AxiosMinJs,
	Oauth2Html,
}
```

- [ ] **Step 3: 取回并转换 icons**

```bash
cd /f/baidu/aiib-go/knife4go && mkdir -p ui/icons && for f in ANDROID_CHROME_192X192_PNG ANDROID_CHROME_512X512_PNG APPLE_TOUCH_ICON_120X120_PNG APPLE_TOUCH_ICON_152X152_PNG APPLE_TOUCH_ICON_180X180_PNG APPLE_TOUCH_ICON_60X60_PNG APPLE_TOUCH_ICON_76X76_PNG APPLE_TOUCH_ICON_PNG FAVICON_16X16_PNG FAVICON_32X32_PNG FAVICON_ICO MSAPPLICATION_ICON_144X144_PNG MSTILE_150X150_PNG SAFARI_PINNED_TAB_SVG; do git show ca6aa81:internal/ui/knife/img/icons/$f.go > ui/icons/$f.go; done
```

转换同 fonts：`ui.MustHex`、`ContentType: ""`（原调用 GetOther）。注意路径形如 `/favicon.ico`（在站点根，无 `webjars` 前缀，以各文件 RELATIVE_PATH 常量为准）。

- [ ] **Step 4: 创建 `ui/icons/assets.go`**

```go
// Package icons 提供 knife4go UI 使用的站点图标资源。
package icons

import "github.com/jasonlabz/knife4go/ui"

// Assets 是 icons 分类的全部静态资源。
var Assets = []ui.Asset{
	AndroidChrome192x192Png,
	AndroidChrome512x512Png,
	AppleTouchIcon120x120Png,
	AppleTouchIcon152x152Png,
	AppleTouchIcon180x180Png,
	AppleTouchIcon60x60Png,
	AppleTouchIcon76x76Png,
	AppleTouchIconPng,
	Favicon16x16Png,
	Favicon32x32Png,
	FaviconICO,
	MsapplicationIcon144x144Png,
	Mstile150x150Png,
	SafariPinnedTabSvg,
}
```

- [ ] **Step 5: 删除 internal/ 残留与调试文件**

```bash
cd /f/baidu/aiib-go/knife4go && rm -rf internal && git add -A
```

（原 `internal/ui/knife/img/icons/file.js`、`hex_test.go` 为生成调试残留，不迁移、不恢复。）

- [ ] **Step 6: register.go 加入 oauth 与 icons**

```go
import (
	// ...现有 import
	"github.com/jasonlabz/knife4go/ui/icons"
	"github.com/jasonlabz/knife4go/ui/webjars/oauth"
)

// allAssets 返回全部静态资产。
func allAssets() []ui.Asset {
	assets := append([]ui.Asset{}, css.Assets...)
	assets = append(assets, js.Assets...)
	assets = append(assets, fonts.Assets...)
	assets = append(assets, img.Assets...)
	assets = append(assets, oauth.Assets...)
	return append(assets, icons.Assets...)
}
```

- [ ] **Step 7: register_test.go 追加 oauth 与 icons 期望路径**

```go
		"/webjars/oauth/axios.min.js",
		"/webjars/oauth/oauth2.html",
		"/android-chrome-192x192.png",
		"/android-chrome-512x512.png",
		"/apple-touch-icon-120x120.png",
		"/apple-touch-icon-152x152.png",
		"/apple-touch-icon-180x180.png",
		"/apple-touch-icon-60x60.png",
		"/apple-touch-icon-76x76.png",
		"/apple-touch-icon.png",
		"/favicon-16x16.png",
		"/favicon-32x32.png",
		"/favicon.ico",
		"/msapplication-icon-144x144.png",
		"/mstile-150x150.png",
		"/safari-pinned-tab.svg",
```

（路径以各文件 RELATIVE_PATH 常量为准。）

- [ ] **Step 8: 验证与提交**

```bash
cd /f/baidu/aiib-go/knife4go && go build ./... && go test ./... && go vet ./... && gofmt -l . && git add -A && git commit -m "feat: data-driven oauth and icons assets, drop internal layer"
```

---

### Task 6: huma 适配器

**Files:**
- Create: `huma/huma.go`、`huma/huma_test.go`

**Interfaces:**
- Consumes: 根包 `Router`、`Register`、`Opts`、`Doc`、`DocPath`
- Produces: `huma.InitSwaggerKnife(api huma.API, opts ...knife.Opts) error`

- [ ] **Step 1: 创建 `huma/huma.go`**

```go
// Package huma 提供 knife4go 的 Huma v2 适配器。
package huma

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	knife "github.com/jasonlabz/knife4go"
)

// humaRouter 将 huma.API 适配为 knife.Router。
// 静态路由以 Hidden 操作注册，不进入 huma 生成的 OpenAPI 文档。
type humaRouter struct {
	api huma.API
}

// GET 实现 knife.Router。
func (w *humaRouter) GET(path, contentType string, content []byte) {
	w.api.Handle(&huma.Operation{
		Method: http.MethodGet,
		Path:   path,
		Hidden: true,
	}, func(ctx huma.Context) {
		ctx.SetStatus(http.StatusOK)
		if contentType != "" {
			ctx.SetHeader("Content-Type", contentType)
		}
		_, _ = ctx.BodyWriter().Write(content)
	})
}

// InitSwaggerKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 huma API。
// 文档来源默认 swag.ReadDoc("swagger")，也可通过 knife.Doc(doc) 显式指定；
// UI 页面路径默认 /doc.html，可用 knife.DocPath(path) 修改。
// 传入 huma.NewGroup(api, prefix) 创建的组可整体加前缀。
func InitSwaggerKnife(api huma.API, opts ...knife.Opts) error {
	return knife.Register(&humaRouter{api: api}, opts...)
}
```

- [ ] **Step 2: 创建 `huma/huma_test.go`**

使用 `humatest.New(tb, configs ...huma.Config) (http.Handler, humatest.TestAPI)`（TestAPI 内嵌 huma.API）：

```go
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
```

- [ ] **Step 3: 验证与提交**

Task 1 的 `go mod tidy` 会因 huma 尚未被引用而移除该依赖；本任务先 `go mod tidy` 重新解析（会重新加入 huma 与传递依赖），再构建测试：

```bash
cd /f/baidu/aiib-go/knife4go && go mod tidy && go build ./... && go test ./... && go vet ./... && gofmt -l . && git add -A && git commit -m "feat: huma adapter"
```

---

### Task 7: _examples 与 README

**Files:**
- Create: `_examples/gin/main.go`、`_examples/huma/main.go`
- Rewrite: `README.md`

- [ ] **Step 1: 创建 `_examples/gin/main.go`**

```go
// Command gin 演示 knife4go 在 gin 框架下的最小接入方式。
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/knife4go"
)

// InitApiRouter 返回带 knife4go UI 的路由。
// 文档默认来自 swag.ReadDoc("swagger")（需空白导入 _ "项目/docs/swagger"），
// 也可用 knife4go.Doc(doc) 显式指定。
func InitApiRouter() *gin.Engine {
	router := gin.Default()
	serverGroup := router.Group("/demo")
	_ = knife4go.InitSwaggerKnife(serverGroup)
	return router
}

func main() {
	_ = InitApiRouter()
}
```

- [ ] **Step 2: 创建 `_examples/huma/main.go`**

```go
// Command huma 演示 knife4go 在 huma v2 框架下的最小接入方式。
package main

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/knife4go/huma"
)

func main() {
	router := gin.Default()
	api := humagin.New(router, huma.DefaultConfig("demo", "v1"))
	_ = huma.InitSwaggerKnife(api)
	_ = router.Run(":8080")
}
```

- [ ] **Step 3: 重写 `README.md`**

内容要点（中文）：项目定位（Knife4j UI + OpenAPI 文档的 Go 库）；支持 gin（根包 `InitSwaggerKnife`，与旧版一致）与 huma（`huma/` 子包）；文档来源默认 swag 或 `Doc()` 指定；`DocPath()` 自定义页面路径；扩展新框架指南（实现 `knife4go.Router` 接口约 15 行，参照 `huma/` 包）；HTTP 端点说明（`/doc.html`、`/v3/api-docs`、`/v3/api-docs/swagger-config`）；go.mod 依赖与 Go 版本要求（Go 1.25）。

- [ ] **Step 4: 验证与提交**

```bash
cd /f/baidu/aiib-go/knife4go && go build ./... && go test ./... && go vet ./... && gofmt -l . && git add -A && git commit -m "docs: rewrite README, add gin and huma examples"
```

---

### Task 8: generate-example-project 调用点迁移

**Files:**
- Modify: `../generate-example-project/server/router/router.go:38-46`、`../generate-example-project/go.mod`（knife4go 版本）

**Interfaces:**
- Consumes: 根包 `InitSwaggerKnife`、`Doc`

- [ ] **Step 1: 迁移 router.go**

将 `generate-example-project/server/router/router.go` 中：

```go
	// Knife4go must observe the completed Huma document when it registers debug routes.
	if debug {
		if err := knife.Init(serverGroup, knife.DocumentProviderFunc(func() ([]byte, error) {
			return routerApi.OpenAPI().Downgrade()
		})); err != nil {
			return nil, fmt.Errorf("initialize knife4go documentation: %w", err)
		}
	}
```

改为：

```go
	// Knife4go must observe the completed Huma document when it registers debug routes.
	if debug {
		openAPIDocument, err := routerApi.OpenAPI().Downgrade()
		if err != nil {
			return nil, fmt.Errorf("downgrade huma OpenAPI document: %w", err)
		}
		if err := knife.InitSwaggerKnife(serverGroup, knife.Doc(string(openAPIDocument))); err != nil {
			return nil, fmt.Errorf("initialize knife4go documentation: %w", err)
		}
	}
```

- [ ] **Step 2: 更新 go.mod 依赖**

在 `generate-example-project` 下验证 knife4go 新版本可解析（若 knife4go 已推送，`go get github.com/jasonlabz/knife4go@<新提交>`；未推送则添加本地替换并随后移除）：

```bash
cd /f/baidu/aiib-go/generate-example-project && go mod edit -replace github.com/jasonlabz/knife4go=../knife4go && go build ./... && go test ./... && go mod tidy
```

验证通过后，若 knife4go 已推送新提交，用 `go get` 替换 replace 指令；若未推送，保留 replace 并提示用户推送后移除。

- [ ] **Step 3: 提交（generate-example-project 仓库由用户决定是否一并提交）**

```bash
cd /f/baidu/aiib-go/generate-example-project && git add server/router/router.go go.mod go.sum && git commit -m "refactor: migrate to knife4go InitSwaggerKnife API"
```

若 generate-example-project 不在独立 git 仓库或用户希望暂缓提交，跳过本步并在最终汇报中说明。

---

## 验证总览（全部任务完成后）

```bash
cd /f/baidu/aiib-go/knife4go && go build ./... && go test -cover ./... && go vet ./... && gofmt -l . && git status --short
```

- 根包覆盖率不低于 80%，huma 包覆盖率不低于 80%。
- 目录结构最终为：`knife.go`、`gin.go`、`register.go`、`opts.go`、`knife_test.go`、`register_test.go`、`ui/`（assets.go、DOC_HTML.go、icons/、webjars/{css,fonts,img,js,oauth}/）、`huma/`、`_examples/`、`docs/`。
- `internal/`、`constant/`、`utils/` 不存在。
- dagine、dagine-dashboard 无需改动（根包 API 未变）。

## 自审记录

- **Spec 覆盖**：回退 API 形态 → Task 1/8；gin+huma+扩展接口 → Task 1/6；去 internal 层 → Task 1-5；资产数据化 → Task 2-5；README/示例 → Task 7；generate-example-project 迁移 → Task 8；swagger-config 计算与 basePath → Task 1；错误处理（无 log.Fatal）→ Task 1。
- **占位符扫描**：无 TBD；资产内容引用原文件常量，转换规则逐类给出完整示例。
- **类型一致性**：`ui.Asset{Path, ContentType, Content}`、`Router.GET(path, contentType string, content []byte)`、`allAssets() []ui.Asset`、`css/js/fonts/img/oauth/icons.Assets`、`InitSwaggerKnife` 双入口签名在各任务间一致。
