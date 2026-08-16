// Package huma 提供 knife4go 的 Huma v2 适配器。
package huma

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	knife "github.com/jasonlabz/knife4go"
)

// humaRouter 将 huma.API 适配为 knife.Router。
// 静态路由以 Hidden 操作注册，不进入 huma 生成的 OpenAPI 文档。
type humaRouter struct {
	api huma.API
}

// GET 实现 knife.Router。
func (w *humaRouter) GET(path, contentType string, content []byte) {
	// huma v2.39.1 中 Handle 定义在 huma.Adapter 上（经 API.Adapter() 获取）；
	// 对 huma.Group 而言 Adapter() 返回组适配器，会自动应用前缀，故传组也能整体加前缀。
	w.api.Adapter().Handle(&huma.Operation{
		Method: http.MethodGet,
		Path:   path,
		Hidden: true,
	}, func(ctx huma.Context) {
		ctx.SetStatus(http.StatusOK)
		if contentType != "" {
			ctx.SetHeader("Content-Type", contentType)
		}
		_, _ = ctx.BodyWriter().Write(content)
	})
}

// InitSwaggerKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 huma API。
// 文档来源默认 swag.ReadDoc("swagger")，也可通过 knife.Doc(doc) 显式指定；
// UI 页面路径默认 /doc.html，可用 knife.DocPath(path) 修改。
//
// 注册位置需与文档 paths 的前缀互补：Knife4j 前端会基于 doc.html 所在路径
// 拼接接口路径（huma.NewGroup 会把组前缀写入文档 paths）。因此 huma 文档
// paths 带前缀时应注册在根级 API；仅当文档 paths 无前缀（如 swag 生成）时
// 才适合注册在带前缀的组上。
func InitSwaggerKnife(api huma.API, opts ...knife.Opts) error {
	return knife.Register(&humaRouter{api: api}, opts...)
}
