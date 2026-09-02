package knife

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/smarteng/knife4go/ui"
)

// Router 是 knife4go 注册静态 GET 路由所需的最小框架无关接口。
// 新增 web/api 框架支持：实现该接口后调用 RegisterOpenAPI，或参照 gin/、huma/
// 子包封装便捷入口。
type Router interface {
	// GET 注册一条 GET 路由，响应体固定为 content，Content-Type 为 contentType（空串表示不设置）。
	GET(path, contentType string, content []byte)
}

const (
	// defaultAPIDocsPath 是 OpenAPI 文档 JSON 端点的默认路径，可经 APIDocsPath 定制。
	defaultAPIDocsPath = "/swagger/doc.json"
	// swaggerConfigEndpoint 是 Knife4j 前端约定的 UI 能力配置端点路径。
	// 该路径由前端 JS 硬编码探测，独立于文档 JSON 端点，不随 APIDocsPath 变化。
	swaggerConfigEndpoint = "/v3/api-docs/swagger-config"
	// oauth2RedirectPath 是 oauth2 重定向页面的相对路径。
	oauth2RedirectPath = "/swagger-ui/oauth2-redirect.html"
	// jsonContentType 是 JSON 响应使用的 Content-Type。
	jsonContentType = "application/json;charset=UTF-8"
)

// errNoDocument 表示调用方未通过 Doc() 提供 OpenAPI 文档内容。
var errNoDocument = errors.New("knife4go: no OpenAPI document provided, use knife.Doc(doc) to supply the OpenAPI 3.0 JSON")

// normalizeAbsPath 规范化路由路径：若为空则回退到 fallback，否则确保以 "/" 开头。
func normalizeAbsPath(p, fallback string) string {
	if p == "" {
		return fallback
	}
	if p[0] != '/' {
		return "/" + p
	}
	return p
}

// uiPrefixOf 从 UI 页面路径（如 /swagger/index.html、/doc.html）推导出用于其它资源的前缀。
// 规则：取所在目录，并去掉末尾的斜杠；根目录返回空串。
//
//	/doc.html               -> ""
//	/swagger/index.html     -> "/swagger"
//	/a/b/index.html         -> "/a/b"
//	/                       -> ""
func uiPrefixOf(uiPath string) string {
	dir := path.Dir(uiPath)
	if dir == "/" || dir == "." {
		return ""
	}
	return strings.TrimRight(dir, "/")
}

// withPrefix 把 uiPrefix 拼到 p 前面。约束：p 必须以 "/" 开头。
// 若 p 已经以 uiPrefix 开头（后接 "/" 或恰好等于 uiPrefix），则原样返回，
// 避免用户显式带前缀时被二次拼接（如 APIDocsPath("/swagger/doc.json")）。
func withPrefix(uiPrefix, p string) string {
	if uiPrefix == "" {
		return p
	}
	// 手工比较代替 strings.HasPrefix(p, uiPrefix+"/") 以避免临时字符串分配。
	if len(p) >= len(uiPrefix) && p[:len(uiPrefix)] == uiPrefix {
		if len(p) == len(uiPrefix) || p[len(uiPrefix)] == '/' {
			return p
		}
	}
	return uiPrefix + p
}

// defaultGroupName 是单文档模式下 urls 数组唯一条目的显示名，
// 会在 knife4j 首页 #/home 的分组链接列表中呈现。
const defaultGroupName = "default"

// buildSwaggerConfigJSON 生成 swagger-config 端点响应体。
//
// 使用 urls 数组形式（而非单值 url 字段）承载文档 JSON 地址，这样即便只有一个
// 文档也能让 knife4j 首页 #/home 渲染出"分组链接"UI，与多分组体验保持一致。
// 内部字段均使用 strip 掉 uiPrefix 后的相对路径——因为 knife4j 前端会基于当前
// UI 页面所在目录（window.location.pathname 的目录部分）自动拼接前缀（a + url）。
func buildSwaggerConfigJSON(apiDocsPath, uiPrefix string) []byte {
	relativeURL := strings.TrimPrefix(apiDocsPath, uiPrefix)
	return []byte(fmt.Sprintf(
		`{"configUrl": %q,"oauth2RedirectUrl": %q,"urls": [{"name": %q,"url": %q,"location": %q}],"validatorUrl": ""}`,
		swaggerConfigEndpoint,
		oauth2RedirectPath,
		defaultGroupName,
		relativeURL,
		relativeURL,
	))
}

// RegisterOpenAPI 将 knife4go 的 UI 页面、OpenAPI 文档端点与全部静态资产注册到 router。
// 文档必须经 Doc() 提供，否则返回错误；文档内容始终原样注册，不做任何改写。
//
// 挂载位置说明：
//   - 用户通过 DocPath 指定 UI 页面路径（默认 /doc.html），knife4go 自动推导目录部分
//     作为 uiPrefix，例如 DocPath("/swagger/index.html") -> uiPrefix="/swagger"。
//   - APIDocsPath、swagger-config、oauth2-redirect、40 条静态资产均自动挂到 uiPrefix
//     下，用户无需重复拼接前缀；APIDocsPath 若已包含 uiPrefix 前缀，则原样使用。
//   - swagger-config 响应体里的 url/configUrl/oauth2RedirectUrl 一律使用 strip 掉
//     uiPrefix 后的相对路径——因为 knife4j 前端会基于当前 UI 页面所在目录自行拼接前缀。
func RegisterOpenAPI(router Router, opts ...Opts) error {
	config := Config{}
	for _, opt := range opts {
		opt(&config)
	}
	if config.docJson == "" {
		return errNoDocument
	}

	// 1) 解析所有路径（UI 页面 + 文档 JSON + uiPrefix）。
	docPath := normalizeAbsPath(config.docPath, ui.DocHtmlRelativePath)
	uiPrefix := uiPrefixOf(docPath)
	apiDocsPath := withPrefix(uiPrefix, normalizeAbsPath(config.apiDocsPath, defaultAPIDocsPath))

	// 2) 动态路由（保留框架路由日志，方便用户确认注册位置）。
	router.GET(docPath, ui.DocHtml.ContentType, ui.DocHtml.Content)
	router.GET(apiDocsPath, jsonContentType, []byte(config.docJson))

	// 3) 静态路由（swagger-config + 40 条 UI 静态资产）。适配器可通过 beforeStaticAssets
	//    钩子在此段前后插入横切逻辑（如临时抑制框架路由日志）。
	registerStaticRoutes(router, uiPrefix, buildSwaggerConfigJSON(apiDocsPath, uiPrefix), config.beforeStaticAssets)
	return nil
}

// registerStaticRoutes 集中注册"静默范围"内的路由：swagger-config 端点与全部 UI 静态资产。
// 若 beforeHook 非 nil，则在开始注册前调用一次，其返回的 cleanup 会在全部注册完成后立即执行。
func registerStaticRoutes(router Router, uiPrefix string, swaggerConfig []byte, beforeHook func() (cleanup func())) {
	var cleanup func()
	if beforeHook != nil {
		cleanup = beforeHook()
	}
	router.GET(withPrefix(uiPrefix, swaggerConfigEndpoint), jsonContentType, swaggerConfig)
	for _, asset := range ui.AllAssets() {
		router.GET(withPrefix(uiPrefix, asset.Path), asset.ContentType, asset.Content)
	}
	if cleanup != nil {
		cleanup()
	}
}
