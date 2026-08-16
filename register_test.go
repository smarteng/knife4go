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
