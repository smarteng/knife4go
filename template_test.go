package knife

import (
	"strings"
	"testing"
)

// TestResolveTemplateDefault 验证不传 Template() 时默认走 "ui" 模板。
func TestResolveTemplateDefault(t *testing.T) {
	p, err := resolveTemplate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.docHtmlRelativePath() != "/doc.html" {
		t.Errorf("expected default template docHtmlRelativePath /doc.html, got %q", p.docHtmlRelativePath())
	}
}

// TestResolveTemplateNamed 验证按名字选择模板。
func TestResolveTemplateNamed(t *testing.T) {
	for _, name := range []string{"ui", "vue"} {
		t.Run(name, func(t *testing.T) {
			p, err := resolveTemplate(name)
			if err != nil {
				t.Fatalf("resolveTemplate(%q) failed: %v", name, err)
			}
			if p == nil {
				t.Fatalf("resolveTemplate(%q) returned nil provider", name)
			}
			// 每个模板都应至少能提供 doc.html
			h := p.docHtml()
			if len(h.Content) == 0 {
				t.Errorf("template %q docHtml content is empty", name)
			}
		})
	}
}

// TestResolveTemplateUnknownReturnsError 验证未知名字返回错误。
func TestResolveTemplateUnknownReturnsError(t *testing.T) {
	_, err := resolveTemplate("unknown-tpl")
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
	if !strings.Contains(err.Error(), "unknown-tpl") {
		t.Errorf("expected error to mention unknown template name, got %q", err.Error())
	}
}

// TestRegisterOpenAPIUsesTemplateOption 验证 Template("vue") 选项能改变 register 的资产来源。
func TestRegisterOpenAPIUsesTemplateOption(t *testing.T) {
	routerUI := &fakeRouter{}
	if err := RegisterOpenAPI(routerUI, Doc(openAPI303Document), Template("ui")); err != nil {
		t.Fatalf("register with ui template failed: %v", err)
	}
	routerVue := &fakeRouter{}
	if err := RegisterOpenAPI(routerVue, Doc(openAPI303Document), Template("vue")); err != nil {
		t.Fatalf("register with vue template failed: %v", err)
	}
	// 两组路由数量可能不同 (不同前端构建产物不同), 但都应包含核心动态路由。
	for _, want := range []string{"/doc.html", "/swagger/doc.json", "/v3/api-docs/swagger-config"} {
		if !containsRoute(routerUI.routes, want) {
			t.Errorf("ui template: expected route %q to be registered", want)
		}
		if !containsRoute(routerVue.routes, want) {
			t.Errorf("vue template: expected route %q to be registered", want)
		}
	}
}

// TestRegisterOpenAPIUnknownTemplateReturnsError 验证传入未知模板名时 RegisterOpenAPI 返回错误。
func TestRegisterOpenAPIUnknownTemplateReturnsError(t *testing.T) {
	router := &fakeRouter{}
	err := RegisterOpenAPI(router, Doc(openAPI303Document), Template("does-not-exist"))
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
}
