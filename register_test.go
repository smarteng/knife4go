package knife

import (
	"bytes"
	"encoding/json"
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
	if err := RegisterOpenAPI(router, append([]Opts{Doc(openAPI303Document)}, opts...)...); err != nil {
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

func TestRegisterRegistersAllStaticAssets(t *testing.T) {
	routes := fakeRouterRoutes(t)
	// 期望路径以各资产文件 RELATIVE_PATH 常量为准（此处与常量一致）
	for _, want := range []string{
		"/webjars/css/app.ac23e017.css",
		"/webjars/css/chunk-75464e7e.8fb93ba5.css",
		"/webjars/css/chunk-d7d5f59c.a9ffbfcb.css",
		"/webjars/css/chunk-vendors.f24a310a.css",
		"/webjars/css/chunk-vendors.f24a310a.css.gz",
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
		"/webjars/fonts/fontawesome-webfont.706450d7.ttf",
		"/webjars/fonts/fontawesome-webfont.97493d3f.woff2",
		"/webjars/fonts/fontawesome-webfont.d9ee23d5.woff",
		"/webjars/fonts/fontawesome-webfont.f7c2b4b7.eot",
		"/webjars/fonts/iconfont.4ca3d0c0.ttf",
		"/webjars/fonts/iconfont.e2d2b98e.eot",
		"/webjars/img/editormd-logo.53ea80e2.svg",
		"/webjars/img/fontawesome-webfont.29800836.svg",
		"/webjars/img/iconfont.1d48c203.svg",
		"/webjars/img/loading2x.695405a9.gif",
		"/webjars/img/loading3x.65eacf61.gif",
		"/webjars/img/loading.c929501e.gif",
		"/webjars/oauth/axios.min.js",
		"/webjars/oauth/oauth2.html",
		"/img/icons/android-chrome-192x192.png",
		"/img/icons/android-chrome-512x512.png",
		"/img/icons/apple-touch-icon-120x120.png",
		"/img/icons/apple-touch-icon-152x152.png",
		"/img/icons/apple-touch-icon-180x180.png",
		"/img/icons/apple-touch-icon-60x60.png",
		"/img/icons/apple-touch-icon-76x76.png",
		"/img/icons/apple-touch-icon.png",
		"/img/icons/favicon-16x16.png",
		"/img/icons/favicon-32x32.png",
		"/favicon.ico",
		"/img/icons/msapplication-icon-144x144.png",
		"/img/icons/mstile-150x150.png",
		"/img/icons/safari-pinned-tab.svg",
	} {
		if !containsRoute(routes, want) {
			t.Errorf("expected asset route %q to be registered", want)
		}
	}
}

func TestRegisterConfigContentUsesPrefixFreeURLs(t *testing.T) {
	router := &fakeRouter{}
	if err := RegisterOpenAPI(router, Doc(openAPI303Document)); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	// Knife4j 前端会基于 doc.html 所在路径自行拼接前缀（a + url），
	// 因此 config 内容必须是无前缀相对路径。
	config := string(routeContent(router.routes, "/v3/api-docs/swagger-config"))
	for _, want := range []string{
		`"configUrl": "/v3/api-docs/swagger-config"`,
		`"oauth2RedirectUrl": "/swagger-ui/oauth2-redirect.html"`,
		`"url": "/v3/api-docs"`,
	} {
		if !strings.Contains(config, want) {
			t.Errorf("expected config to contain %s, got %q", want, config)
		}
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

const prefixedOpenAPIDocument = `{"openapi":"3.0.3","info":{"title":"t","version":"1.0.0"},"paths":{"/catalog/health-check":{"get":{}},"/catalog/api/v1/users":{"get":{}}}}`

func TestRegisterOpenAPIStripsDocumentBasePath(t *testing.T) {
	router := &fakeRouter{}
	if err := RegisterOpenAPI(router, Doc(prefixedOpenAPIDocument), DocBasePath("/catalog")); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	doc := routeContent(router.routes, "/v3/api-docs")
	var root struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(doc, &root); err != nil {
		t.Fatalf("unexpected document: %v", err)
	}
	for _, want := range []string{"/health-check", "/api/v1/users"} {
		if _, ok := root.Paths[want]; !ok {
			t.Errorf("expected path %q after stripping base path, got %v", want, keys(root.Paths))
		}
	}
	for _, unwanted := range []string{"/catalog/health-check", "/catalog/api/v1/users"} {
		if _, ok := root.Paths[unwanted]; ok {
			t.Errorf("expected path %q to be stripped, got %v", unwanted, keys(root.Paths))
		}
	}
}

func TestRegisterOpenAPIKeepsDocumentWhenBasePathDoesNotMatch(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		doc   string
		opts  []Opts
		paths []string
	}{
		{name: "unprefixed document with base path option", doc: openAPI303Document, opts: []Opts{DocBasePath("/catalog")}, paths: []string{}},
		{name: "prefixed document without base path option", doc: prefixedOpenAPIDocument, opts: nil, paths: []string{"/catalog/health-check"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			router := &fakeRouter{}
			if err := RegisterOpenAPI(router, append([]Opts{Doc(testCase.doc)}, testCase.opts...)...); err != nil {
				t.Fatalf("unexpected registration error: %v", err)
			}
			doc := routeContent(router.routes, "/v3/api-docs")
			var root struct {
				Paths map[string]json.RawMessage `json:"paths"`
			}
			if err := json.Unmarshal(doc, &root); err != nil {
				t.Fatalf("unexpected document: %v", err)
			}
			for _, want := range testCase.paths {
				if _, ok := root.Paths[want]; !ok {
					t.Errorf("expected path %q preserved as-is, got %v", want, keys(root.Paths))
				}
			}
		})
	}
}

func keys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
