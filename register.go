package knife

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/swaggo/swag"

	"github.com/jasonlabz/knife4go/ui"
	"github.com/jasonlabz/knife4go/ui/icons"
	"github.com/jasonlabz/knife4go/ui/webjars/css"
	"github.com/jasonlabz/knife4go/ui/webjars/fonts"
	"github.com/jasonlabz/knife4go/ui/webjars/img"
	"github.com/jasonlabz/knife4go/ui/webjars/js"
	"github.com/jasonlabz/knife4go/ui/webjars/oauth"
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

// RegisterOpenAPI 将 knife4go 的 UI 页面、OpenAPI 文档端点与全部静态资产注册到 router。
func RegisterOpenAPI(router Router, opts ...Opts) error {
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

	// UI 页面（路径可经 DocPath 定制）
	docPath := config.docPath
	if docPath == "" {
		docPath = ui.DocHtmlRelativePath
	}
	router.GET(docPath, ui.DocHtml.ContentType, ui.DocHtml.Content)

	// 注册位置前缀：DocBasePath 显式声明优先，否则由适配器提供（如 gin 的 BasePath）。
	basePath := config.basePath
	if basePath == "" {
		if routerWithBasePath, ok := router.(interface{ BasePath() string }); ok {
			basePath = routerWithBasePath.BasePath()
		}
	}
	// 文档 paths 若全部带注册位置前缀，注册前剥离：Knife4j 前端会基于
	// doc.html 所在路径自行拼接前缀，剥离后拼接结果恰好是原始路径。
	docJSON := config.docJson
	if stripped, ok := stripDocumentBasePath([]byte(config.docJson), basePath); ok {
		docJSON = string(stripped)
	}

	// OpenAPI 文档端点与 UI 配置端点均以相对路径注册，由适配器/框架拼接自身前缀。
	// swagger-config 内的 url/configUrl/oauth2RedirectUrl 必须是无前缀相对路径：
	// Knife4j 前端（knife4j-vue）会基于 doc.html 所在路径推导前缀并自行拼接（a + url）。
	configContent := fmt.Sprintf(`{"configUrl": %q,"oauth2RedirectUrl": %q,"url": %q,"validatorUrl": ""}`,
		apiDocsPath+swaggerConfigSuffix,
		oauth2RedirectPath,
		apiDocsPath,
	)
	router.GET(apiDocsPath, jsonContentType, []byte(docJSON))
	router.GET(apiDocsPath+swaggerConfigSuffix, jsonContentType, []byte(configContent))

	// 静态资产
	for _, asset := range allAssets() {
		router.GET(asset.Path, asset.ContentType, asset.Content)
	}
	return nil
}

// stripDocumentBasePath 在文档全部 paths 都以 basePath 开头时剥离该前缀并返回 true；
// 否则返回原始文档与 false。仅改写 paths 键，其余字段原样保留。
func stripDocumentBasePath(document []byte, basePath string) ([]byte, bool) {
	basePath = strings.TrimSuffix(basePath, "/")
	if basePath == "" {
		return document, false
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(document, &root); err != nil {
		return document, false
	}
	pathsRaw, ok := root["paths"]
	if !ok {
		return document, false
	}
	var paths map[string]json.RawMessage
	if err := json.Unmarshal(pathsRaw, &paths); err != nil || len(paths) == 0 {
		return document, false
	}
	for p := range paths {
		if !strings.HasPrefix(p, basePath+"/") {
			return document, false
		}
	}
	stripped := make(map[string]json.RawMessage, len(paths))
	for p, v := range paths {
		stripped[strings.TrimPrefix(p, basePath)] = v
	}
	newPaths, err := json.Marshal(stripped)
	if err != nil {
		return document, false
	}
	root["paths"] = newPaths
	out, err := json.Marshal(root)
	if err != nil {
		return document, false
	}
	return out, true
}

// allAssets 返回全部静态资产。
func allAssets() []ui.Asset {
	assets := append([]ui.Asset{}, css.Assets...)
	assets = append(assets, js.Assets...)
	assets = append(assets, fonts.Assets...)
	assets = append(assets, img.Assets...)
	assets = append(assets, oauth.Assets...)
	return append(assets, icons.Assets...)
}
