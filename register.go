package knife

import (
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
	// apiDocsPrefix 是 UI 能力配置（swagger-config）端点的固定前缀。
	// 该前缀由 Knife4j 前端约定，独立于文档 JSON 端点，不随 APIDocsPath 变化。
	apiDocsPrefix = "/v3/api-docs"
	// swaggerConfigSuffix 是 UI 配置端点的相对路径后缀。
	swaggerConfigSuffix = "/swagger-config"
	// oauth2RedirectPath 是 oauth2 重定向页面的相对路径。
	oauth2RedirectPath = "/swagger-ui/oauth2-redirect.html"
	// jsonContentType 是 JSON 响应使用的 Content-Type。
	jsonContentType = "application/json;charset=UTF-8"
)

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

// withPrefix 将 uiPrefix 拼到 p 前面；p 保证以 "/" 开头。若 p 已经以 prefix 开头则保持不变。
func withPrefix(uiPrefix, p string) string {
	if uiPrefix == "" {
		return p
	}
	if strings.HasPrefix(p, uiPrefix+"/") || p == uiPrefix {
		return p
	}
	return uiPrefix + p
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
		return fmt.Errorf("no OpenAPI document provided: use knife.Doc(doc) to supply the OpenAPI 3.0 JSON")
	}

	// UI 页面（路径可经 DocPath 定制），其目录部分推导为 uiPrefix。
	docPath := config.docPath
	if docPath == "" {
		docPath = ui.DocHtmlRelativePath
	}
	uiPrefix := uiPrefixOf(docPath)
	router.GET(docPath, ui.DocHtml.ContentType, ui.DocHtml.Content)

	// OpenAPI 文档 JSON 端点：默认自动挂到 uiPrefix 下；用户显式传入的 APIDocsPath
	// 若已包含 uiPrefix 前缀则保持原样，否则自动补齐前缀。
	apiDocsPath := config.apiDocsPath
	if apiDocsPath == "" {
		apiDocsPath = defaultAPIDocsPath
	}
	apiDocsPath = withPrefix(uiPrefix, apiDocsPath)

	// swagger-config 响应体里的路径必须"去除 uiPrefix"——因为 knife4j 前端会基于
	// window.location.pathname 的目录部分自动拼接前缀（a + url）。
	relativeAPIDocsPath := strings.TrimPrefix(apiDocsPath, uiPrefix)
	configContent := fmt.Sprintf(`{"configUrl": %q,"oauth2RedirectUrl": %q,"url": %q,"validatorUrl": ""}`,
		apiDocsPrefix+swaggerConfigSuffix,
		oauth2RedirectPath,
		relativeAPIDocsPath,
	)
	router.GET(apiDocsPath, jsonContentType, []byte(config.docJson))

	// 静态资产（40 条固定路由）与 swagger-config 端点—— 适配器可通过
	// beforeStaticAssets 钩子在此段前后插入横切逻辑（如临时抑制框架路由日志），
	// 仅覆盖此段批量注册，不影响前面 2 条动态路由（doc.html、apiDocsPath）的日志。
	var cleanup func()
	if config.beforeStaticAssets != nil {
		cleanup = config.beforeStaticAssets()
	}
	router.GET(withPrefix(uiPrefix, apiDocsPrefix+swaggerConfigSuffix), jsonContentType, []byte(configContent))
	for _, asset := range ui.AllAssets() {
		router.GET(withPrefix(uiPrefix, asset.Path), asset.ContentType, asset.Content)
	}
	if cleanup != nil {
		cleanup()
	}
	return nil
}
