package knife

import (
	"fmt"

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
	// apiDocsPath 是 OpenAPI 文档的相对路径。
	apiDocsPath = "/swagger/doc.json"
	// swaggerConfigSuffix 是 UI 配置端点的相对路径后缀。
	// swaggerConfigSuffix = "/swagger-config"
	// oauth2RedirectPath 是 oauth2 重定向页面的相对路径。
	// oauth2RedirectPath = "/swagger-ui/oauth2-redirect.html"
	// jsonContentType 是 JSON 响应使用的 Content-Type。
	jsonContentType = "application/json;charset=UTF-8"
)

// RegisterOpenAPI 将 knife4go 的 UI 页面、OpenAPI 文档端点与全部静态资产注册到 router。
// 文档必须经 Doc() 提供，否则返回错误；文档内容始终原样注册，不做任何改写。
// 前端按文档 paths 是否已带注册位置前缀自行决定是否拼接 doc.html 位置前缀
// （huma 组前缀写入 paths 时不再拼接）。
func RegisterOpenAPI(router Router, opts ...Opts) error {
	config := Config{}
	for _, opt := range opts {
		opt(&config)
	}
	if config.docJson == "" {
		return fmt.Errorf("no OpenAPI document provided: use knife.Doc(doc) to supply the OpenAPI 3.0 JSON")
	}

	// UI 页面（路径可经 DocPath 定制）
	docPath := config.docPath
	if docPath == "" {
		docPath = ui.DocHtmlRelativePath
	}
	router.GET(docPath, ui.DocHtml.ContentType, ui.DocHtml.Content)

	// OpenAPI 文档端点与 UI 配置端点均以相对路径注册，由适配器/框架拼接自身前缀。
	// swagger-config 内的 url/configUrl/oauth2RedirectUrl 必须是无前缀相对路径：
	// Knife4j 前端（knife4j-vue）会基于 doc.html 所在路径推导前缀并自行拼接（a + url）。
	// configContent := fmt.Sprintf(`{"configUrl": %q,"oauth2RedirectUrl": %q,"url": %q,"validatorUrl": ""}`,
	// 	apiDocsPath+swaggerConfigSuffix,
	// 	oauth2RedirectPath,
	// 	apiDocsPath,
	// )
	router.GET(apiDocsPath, jsonContentType, []byte(config.docJson))
	// router.GET(apiDocsPath+swaggerConfigSuffix, jsonContentType, []byte(configContent))

	// 静态资产
	for _, asset := range ui.AllAssets() {
		router.GET(asset.Path, asset.ContentType, asset.Content)
	}
	return nil
}
