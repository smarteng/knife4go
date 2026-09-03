package knife

import (
	"errors"
	"fmt"
	"path"
	"strings"
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
func buildSwaggerConfigJSON(apiDocsPath, uiPrefix string) []byte {
	return []byte(fmt.Sprintf(
		`{"configUrl": %q,"oauth2RedirectUrl": %q,"urls": [{"name": %q,"url": %q,"location": %q}],"validatorUrl": ""}`,
		swaggerConfigEndpoint,
		oauth2RedirectPath,
		defaultGroupName,
		apiDocsPath,
		apiDocsPath,
	))
}

// RegisterOpenAPI 将 knife4go 的 UI 页面、OpenAPI 文档端点与全部静态资产注册到 router。
// 文档必须经 Doc() 提供，否则返回错误；文档内容始终原样注册，不做任何改写。
// 前端模板可通过 Template(name) 选择, 默认使用 "ui" (Vue2 版)。
func RegisterOpenAPI(router Router, opts ...Opts) error {
	config := Config{}
	for _, opt := range opts {
		opt(&config)
	}
	if config.docJson == "" {
		return errNoDocument
	}

	// 0) 解析前端模板。传入未知名称时提前报错, 避免静默回退到默认屏蔽用户拼写错误。
	provider, err := resolveTemplate(config.template)
	if err != nil {
		return err
	}
	docHTML := provider.docHtml()

	// 1) 解析所有路径（UI 页面 + 文档 JSON + uiPrefix）。
	docPath := normalizeAbsPath(config.docPath, provider.docHtmlRelativePath())
	uiPrefix := uiPrefixOf(docPath)
	apiDocsPath := withPrefix(uiPrefix, normalizeAbsPath(config.apiDocsPath, defaultAPIDocsPath))

	// 2) 动态路由（保留框架路由日志，方便用户确认注册位置）。
	router.GET(docPath, docHTML.ContentType, docHTML.Content)
	router.GET(apiDocsPath, jsonContentType, []byte(config.docJson))

	// 3) 静态路由（swagger-config + UI 静态资产）。适配器可通过 beforeStaticAssets
	//    钩子在此段前后插入横切逻辑（如临时抑制框架路由日志）。
	registerStaticRoutes(router, uiPrefix, buildSwaggerConfigJSON(apiDocsPath, uiPrefix), provider.allAssets(), config.beforeStaticAssets)
	return nil
}

// registerStaticRoutes 集中注册"静默范围"内的路由：swagger-config 端点与全部 UI 静态资产。
// 若 beforeHook 非 nil，则在开始注册前调用一次，其返回的 cleanup 会在全部注册完成后立即执行。
func registerStaticRoutes(router Router, uiPrefix string, swaggerConfig []byte, assets []asset, beforeHook func() (cleanup func())) {
	var cleanup func()
	if beforeHook != nil {
		cleanup = beforeHook()
	}
	router.GET(withPrefix(uiPrefix, swaggerConfigEndpoint), jsonContentType, swaggerConfig)
	for _, a := range assets {
		router.GET(withPrefix(uiPrefix, a.Path), a.ContentType, a.Content)
	}
	if cleanup != nil {
		cleanup()
	}
}
