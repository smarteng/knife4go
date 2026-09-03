// Package knife 前端模板选择：通过 knife.Template(name) 选项按名称选择不同的
// UI 前端实现（如 "ui" / "vue"），由 register.RegisterOpenAPI 消费。
package knife

import (
	"fmt"

	"github.com/smarteng/knife4go/ui"
	"github.com/smarteng/knife4go/vue"
)

// asset 是 register 层需要的最小静态资产描述,
// 与 ui.Asset / vue.Asset 结构对称,用于跨模板包统一处理。
type asset struct {
	Path        string
	ContentType string
	Content     []byte
}

// templateProvider 定义前端模板包必须提供的能力:
// 1) 返回 UI 页面 (doc.html) 资产
// 2) 返回 UI 页面的默认相对路径
// 3) 返回全部其它静态资产列表
// 该接口在包内私有,通过 templateRegistry 按名称查找。
type templateProvider interface {
	docHtml() asset
	docHtmlRelativePath() string
	allAssets() []asset
}

// uiProvider 适配 github.com/smarteng/knife4go/ui 包为 templateProvider。
type uiProvider struct{}

func (uiProvider) docHtml() asset {
	h := ui.DocHtml
	return asset{Path: h.Path, ContentType: h.ContentType, Content: h.Content}
}

func (uiProvider) docHtmlRelativePath() string { return ui.DocHtmlRelativePath }

func (uiProvider) allAssets() []asset {
	src := ui.AllAssets()
	out := make([]asset, len(src))
	for i, a := range src {
		out[i] = asset{Path: a.Path, ContentType: a.ContentType, Content: a.Content}
	}
	return out
}

// vueProvider 适配 github.com/smarteng/knife4go/vue 包为 templateProvider。
type vueProvider struct{}

func (vueProvider) docHtml() asset {
	h := vue.DocHtml
	return asset{Path: h.Path, ContentType: h.ContentType, Content: h.Content}
}

func (vueProvider) docHtmlRelativePath() string { return vue.DocHtmlRelativePath }

func (vueProvider) allAssets() []asset {
	src := vue.AllAssets()
	out := make([]asset, len(src))
	for i, a := range src {
		out[i] = asset{Path: a.Path, ContentType: a.ContentType, Content: a.Content}
	}
	return out
}

// templateRegistry 按名称登记可用的前端模板 provider。
// 新增前端只需在此登记一次,调用方即可通过 knife.Template(name) 使用。
var templateRegistry = map[string]templateProvider{
	"ui":  uiProvider{},
	"vue": vueProvider{},
}

// defaultTemplateName 默认前端模板名。未调用 Template() 时使用。
const defaultTemplateName = "ui"

// resolveTemplate 根据名称查找 provider; 名称为空时返回默认 provider。
// 名称不存在时返回明确错误,由 RegisterOpenAPI 传出。
func resolveTemplate(name string) (templateProvider, error) {
	if name == "" {
		name = defaultTemplateName
	}
	p, ok := templateRegistry[name]
	if !ok {
		return nil, fmt.Errorf("knife4go: unknown template %q, available: %v", name, templateNames())
	}
	return p, nil
}

// templateNames 返回已登记的所有模板名, 供错误信息展示。
func templateNames() []string {
	names := make([]string, 0, len(templateRegistry))
	for n := range templateRegistry {
		names = append(names, n)
	}
	return names
}
