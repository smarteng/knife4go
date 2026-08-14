package knife

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jasonlabz/knife4go/internal/ui"
	"github.com/jasonlabz/knife4go/internal/ui/knife"
	"github.com/jasonlabz/knife4go/internal/ui/knife/img/icons"
	"github.com/jasonlabz/knife4go/internal/ui/knife/webjars/css"
	"github.com/jasonlabz/knife4go/internal/ui/knife/webjars/fonts"
	"github.com/jasonlabz/knife4go/internal/ui/knife/webjars/img"
	"github.com/jasonlabz/knife4go/internal/ui/knife/webjars/js"
	"github.com/jasonlabz/knife4go/internal/ui/knife/webjars/oauth"
)

// DocumentProvider supplies a complete OpenAPI 3.0 JSON document.
type DocumentProvider interface {
	OpenAPI() ([]byte, error)
}

// DocumentProviderFunc adapts a function to DocumentProvider.
type DocumentProviderFunc func() ([]byte, error)

// OpenAPI returns the document provided by f.
func (f DocumentProviderFunc) OpenAPI() ([]byte, error) {
	if f == nil {
		return nil, fmt.Errorf("document provider is nil")
	}
	return f()
}

// Option configures Init.
type Option func(*config)

type config struct {
	docPath string
}

// WithDocPath sets the path that serves the UI HTML document.
func WithDocPath(path string) Option {
	return func(c *config) {
		c.docPath = path
	}
}

// Init registers Knife4go's UI, document endpoint, and static assets.
func Init(router gin.IRouter, provider DocumentProvider, options ...Option) error {
	document, err := resolveOpenAPIDocument(provider)
	if err != nil {
		return err
	}

	config := config{}
	for _, option := range options {
		option(&config)
	}

	ui.AddOpenAPIDocumentRouter(router, "", string(document))
	ui.AddOpenAPIConfigRouter(router, "", openAPIDocumentURL(router))

	knife.AddRouterOfDocHtml(router, config.docPath)

	icons.AddRouterOfAndroidChrome192x192Png(router)
	icons.AddRouterOfAndroidChrome512x512Png(router)
	icons.AddRouterOfAppleTouchIcon120x120Png(router)
	icons.AddRouterOfAppleTouchIcon152x152Png(router)
	icons.AddRouterOfAppleTouchIcon180x180Png(router)
	icons.AddRouterOfAppleTouchIcon60x60Png(router)
	icons.AddRouterOfAppleTouchIcon76x76Png(router)
	icons.AddRouterOfAppleTouchIconPng(router)
	icons.AddRouterOfFavicon16x16Png(router)
	icons.AddRouterOfFavicon32x32Png(router)
	icons.AddRouterOfFaviconICO(router)
	icons.AddRouterOfMsapplicationIcon144x144Png(router)
	icons.AddRouterOfMstile150x150Png(router)
	icons.AddRouterOfSafariPinnedTabSvg(router)

	css.AddRouterOfAppAc23e017Css(router)
	css.AddRouterOfChunk75464e7e8fb93ba5Css(router)
	css.AddRouterOfChunkD7d5f59cA9ffbfcbCss(router)
	css.AddRouterOfChunkVendorsF24a310aCss(router)
	css.AddRouterOfChunkVendorsF24a310aCssGz(router)

	fonts.AddRouterOfFontawesomeWebfont706450d7Ttf(router)
	fonts.AddRouterOfFontawesomeWebfont97493d3fWoff2(router)
	fonts.AddRouterOfFontawesomeWebfontD9ee23d5Woff(router)
	fonts.AddRouterOfFontawesomeWebfontF7c2b4b7Eot(router)
	fonts.AddRouterOfIconfont4ca3d0c0Ttf(router)
	fonts.AddRouterOfIconfontE2d2b98eEot(router)

	img.AddRouterOfEditormdLogo53ea80e2Svg(router)
	img.AddRouterOfFontawesomeWebfont29800836Svg(router)
	img.AddRouterOfIconfont1d48c203Svg(router)
	img.AddRouterOfLoadingC929501eGif(router)
	img.AddRouterOfLoading2x695405a9Gif(router)
	img.AddRouterOfLoading3x65eacf61Gif(router)

	js.AddRouterOfApp2fab4ac5Js(router)
	js.AddRouterOfChunk069eb4372cfebf27Js(router)
	js.AddRouterOfChunk069eb4372cfebf27JsLICENSETxt(router)
	js.AddRouterOfChunk0d102d5aB2bddffcJs(router)
	js.AddRouterOfChunk0fd67716D57e2c41Js(router)
	js.AddRouterOfChunk260d712a390177feJs(router)
	js.AddRouterOfChunk260d712a390177feJsLICENSETxt(router)
	js.AddRouterOfChunk2d0af44e392afcd6Js(router)
	js.AddRouterOfChunk2d0bd799Eb48b7f1Js(router)
	js.AddRouterOfChunk2d0da532591ad7fcJs(router)
	js.AddRouterOfChunk3b888a658737ce4fJs(router)
	js.AddRouterOfChunk3ec4aaa8A79d19f8Js(router)
	js.AddRouterOfChunk589faee05bfd1708Js(router)
	js.AddRouterOfChunk589faee05bfd1708JsLICENSETxt(router)
	js.AddRouterOfChunk735c675c5b409314Js(router)
	js.AddRouterOfChunk75464e7eB130271bJs(router)
	js.AddRouterOfChunkAdb9e9442c7f24feJs(router)
	js.AddRouterOfChunkAdb9e9442c7f24feJsLICENSETxt(router)
	js.AddRouterOfChunkD7d5f59cE61130f3Js(router)
	js.AddRouterOfChunkD7d5f59cE61130f3JsLICENSETxt(router)
	js.AddRouterOfChunkVendorsD51cf6f8Js(router)
	js.AddRouterOfChunkVendorsD51cf6f8JsLICENSETxt(router)

	oauth.AddRouterOfAxiosMinJs(router)
	oauth.AddRouterOfOauth2Html(router)

	return nil
}

// resolveOpenAPIDocument resolves and validates one OpenAPI 3.0 JSON document.
func resolveOpenAPIDocument(provider DocumentProvider) ([]byte, error) {
	if provider == nil {
		return nil, fmt.Errorf("resolve OpenAPI document: provider is nil")
	}

	document, err := provider.OpenAPI()
	if err != nil {
		return nil, fmt.Errorf("resolve OpenAPI document: %w", err)
	}
	if len(bytes.TrimSpace(document)) == 0 {
		return nil, fmt.Errorf("validate OpenAPI document: document is blank")
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(document, &root); err != nil {
		return nil, fmt.Errorf("validate OpenAPI document JSON: %w", err)
	}
	versionJSON, ok := root["openapi"]
	if !ok {
		return nil, fmt.Errorf("validate OpenAPI document: missing openapi version")
	}

	var version string
	if err := json.Unmarshal(versionJSON, &version); err != nil {
		return nil, fmt.Errorf("validate OpenAPI document: invalid openapi version: %w", err)
	}
	if !isOpenAPI30Version(version) {
		return nil, fmt.Errorf("validate OpenAPI document: unsupported version %q", version)
	}

	return document, nil
}

// isOpenAPI30Version reports whether version is an OpenAPI 3.0.x version.
func isOpenAPI30Version(version string) bool {
	patch, ok := strings.CutPrefix(version, "3.0.")
	if !ok || patch == "" {
		return false
	}
	for _, digit := range patch {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

// openAPIDocumentURL returns the document URL registered on router.
func openAPIDocumentURL(router gin.IRouter) string {
	basePath := ""
	if routerWithBasePath, ok := router.(interface{ BasePath() string }); ok {
		basePath = routerWithBasePath.BasePath()
	}
	return strings.TrimSuffix(basePath, "/") + ui.API_DOCS_RELATIVE_PATH
}
