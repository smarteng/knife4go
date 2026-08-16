package knife

import (
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

	// OpenAPI 文档端点与 UI 配置端点均以相对路径注册，由适配器/框架拼接自身前缀；
	// basePath 只进入 swagger-config 内容，保证前端拿到的 url 是完整地址。
	configContent := fmt.Sprintf(`{"configUrl": %q,"oauth2RedirectUrl": %q,"url": %q,"validatorUrl": ""}`,
		basePath+apiDocsPath+swaggerConfigSuffix,
		basePath+oauth2RedirectPath,
		docURL,
	)
	router.GET(apiDocsPath, jsonContentType, []byte(config.docJson))
	router.GET(apiDocsPath+swaggerConfigSuffix, jsonContentType, []byte(configContent))

	// 静态资产
	for _, asset := range allAssets() {
		router.GET(asset.Path, asset.ContentType, asset.Content)
	}
	return nil
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
