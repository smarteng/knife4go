// Package huma 提供 knife4go 的 Huma v2 适配器。
package huma

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Router 将 huma.API 适配为 knife.Router。
// 静态路由以 Hidden 操作注册，不进入 huma 生成的 OpenAPI 文档。
type Router struct {
	API huma.API
}

// GET 实现 knife.Router。
func (w *Router) GET(path, contentType string, content []byte) {
	// huma v2.39.1 中 Handle 定义在 huma.Adapter 上（经 API.Adapter() 获取）；
	// 对 huma.Group 而言 Adapter() 返回组适配器，会自动应用前缀，故传组也能整体加前缀。
	w.API.Adapter().Handle(&huma.Operation{
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
