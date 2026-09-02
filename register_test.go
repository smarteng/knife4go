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
	if err := RegisterOpenAPI(router, append([]Opts{Doc(openAPI303Document)}, opts...)...); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	return router.routes
}

func TestRegisterRegistersDynamicRoutes(t *testing.T) {
	routes := fakeRouterRoutes(t)

	for _, want := range []string{"/doc.html", "/swagger/doc.json", "/v3/api-docs/swagger-config"} {
		if !containsRoute(routes, want) {
			t.Errorf("expected route %q to be registered", want)
		}
	}
	if got := routeContent(routes, "/swagger/doc.json"); string(got) != openAPI303Document {
		t.Errorf("expected /swagger/doc.json to serve the OpenAPI document, got %q", got)
	}
	if got := routeContent(routes, "/v3/api-docs/swagger-config"); !strings.Contains(string(got), `"urls": [{"name": "default","url": "/swagger/doc.json"`) {
		t.Errorf("expected swagger-config to contain the document URL, got %q", got)
	}
	if got := routeContent(routes, "/doc.html"); !bytes.Contains(got, []byte("knife4j-vue")) {
		t.Errorf("expected /doc.html to serve the UI page, got %q", got)
	}
}

func TestRegisterRegistersAllStaticAssets(t *testing.T) {
	routes := fakeRouterRoutes(t)
	for _, want := range []string{
		"/webjars/css/app.ac23e017.css",
		"/webjars/css/chunk-75464e7e.8fb93ba5.css",
		"/webjars/css/chunk-d7d5f59c.a9ffbfcb.css",
		"/webjars/css/chunk-vendors.f24a310a.css",
		"/webjars/css/chunk-vendors.f24a310a.css.gz",
		"/webjars/fonts/fontawesome-webfont.706450d7.ttf",
		"/webjars/fonts/fontawesome-webfont.97493d3f.woff2",
		"/webjars/fonts/fontawesome-webfont.d9ee23d5.woff",
		"/webjars/fonts/fontawesome-webfont.f7c2b4b7.eot",
		"/webjars/fonts/iconfont.4ca3d0c0.ttf",
		"/webjars/fonts/iconfont.e2d2b98e.eot",
		"/webjars/img/editormd-logo.53ea80e2.svg",
		"/webjars/img/fontawesome-webfont.29800836.svg",
		"/webjars/img/iconfont.1d48c203.svg",
		"/webjars/img/loading.c929501e.gif",
		"/webjars/img/loading2x.695405a9.gif",
		"/webjars/img/loading3x.65eacf61.gif",
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
		"/webjars/oauth/axios.min.js",
		"/webjars/oauth/oauth2.html",
	} {
		if !containsRoute(routes, want) {
			t.Errorf("expected asset route %q to be registered", want)
		}
	}
}

func TestRegisterFailsWithoutDocument(t *testing.T) {
	router := &fakeRouter{}
	if err := RegisterOpenAPI(router); err == nil {
		t.Fatal("expected error when no document is provided, got nil")
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
		`"urls": [{"name": "default","url": "/swagger/doc.json","location": "/swagger/doc.json"}]`,
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

// TestUIPrefixOf 覆盖 uiPrefixOf 各类边界输入（根目录、多级目录、末尾斜杠等）。
func TestUIPrefixOf(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"/doc.html", ""},
		{"/swagger/index.html", "/swagger"},
		{"/a/b/index.html", "/a/b"},
		{"/", ""},
		{"/nested/", "/nested"},
		{"", ""},
	} {
		t.Run(tc.in, func(t *testing.T) {
			if got := uiPrefixOf(tc.in); got != tc.want {
				t.Errorf("uiPrefixOf(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestWithPrefix 覆盖 withPrefix 各类前缀拼接情形。
func TestWithPrefix(t *testing.T) {
	for _, tc := range []struct {
		name, prefix, in, want string
	}{
		{"empty prefix keeps path", "", "/webjars/x.js", "/webjars/x.js"},
		{"prefix prepended", "/swagger", "/webjars/x.js", "/swagger/webjars/x.js"},
		{"already prefixed keeps original", "/swagger", "/swagger/webjars/x.js", "/swagger/webjars/x.js"},
		{"equals prefix returns as-is", "/swagger", "/swagger", "/swagger"},
		{"partial-match string is prepended", "/swag", "/swagger/x.js", "/swag/swagger/x.js"},
		{"root prefix (empty) with root path", "", "/", "/"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := withPrefix(tc.prefix, tc.in); got != tc.want {
				t.Errorf("withPrefix(%q, %q) = %q, want %q", tc.prefix, tc.in, got, tc.want)
			}
		})
	}
}

// TestNormalizeAbsPath 覆盖 normalizeAbsPath 的三条主分支：空、缺前导斜杠、正常。
func TestNormalizeAbsPath(t *testing.T) {
	for _, tc := range []struct {
		name, in, fallback, want string
	}{
		{"empty uses fallback", "", "/doc.html", "/doc.html"},
		{"missing leading slash", "swagger/doc.json", "/x", "/swagger/doc.json"},
		{"already absolute stays", "/api/openapi.json", "/x", "/api/openapi.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeAbsPath(tc.in, tc.fallback); got != tc.want {
				t.Errorf("normalizeAbsPath(%q, %q) = %q, want %q", tc.in, tc.fallback, got, tc.want)
			}
		})
	}
}
