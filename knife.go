// Package knife 将 Knife4j UI 与 OpenAPI 文档注册到 gin 路由组，
// 或任何实现了 Router 接口的框架适配器（见 huma 子包）。
package knife

import "github.com/gin-gonic/gin"

// InitSwaggerKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 gin 路由组。
// 文档来源默认 swag.ReadDoc("swagger")（需空白导入 _ "项目/docs/swagger"），
// 也可通过 Doc(doc) 显式指定；UI 页面路径默认 /doc.html，可用 DocPath(path) 修改。
func InitSwaggerKnife(router gin.IRouter, opts ...Opts) error {
	return Register(&ginRouter{r: router}, opts...)
}
