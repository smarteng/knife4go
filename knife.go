// Package knife 将 Knife4j UI 与 OpenAPI 文档注册到 gin 路由组，
// 或任何实现了 Router 接口的框架适配器（见 huma 子包）。
package knife

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"

	_gin "github.com/jasonlabz/knife4go/gin"
	_huma "github.com/jasonlabz/knife4go/huma"
)

// InitSwaggerKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 gin 路由组。
// 文档来源默认 swag.ReadDoc("swagger")（需空白导入 _ "项目/docs/swagger"），
// 也可通过 Doc(doc) 显式指定；UI 页面路径默认 /doc.html，可用 DocPath(path) 修改。
func InitSwaggerKnife(router gin.IRouter, opts ...Opts) error {
	return RegisterOpenAPI(&_gin.Router{R: router}, opts...)
}

// InitHumaKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 huma API。
// 文档来源默认 swag.ReadDoc("swagger")，也可通过 knife.Doc(doc) 显式指定；
// UI 页面路径默认 /doc.html，可用 knife.DocPath(path) 修改。
//
// 注册位置需与文档 paths 的前缀互补：Knife4j 前端会基于 doc.html 所在路径
// 拼接接口路径（huma.NewGroup 会把组前缀写入文档 paths）。因此 huma 文档
// paths 带前缀时应注册在根级 API；仅当文档 paths 无前缀（如 swag 生成）时
// 才适合注册在带前缀的组上。
func InitHumaKnife(api huma.API, opts ...Opts) error {
	return RegisterOpenAPI(&_huma.Router{API: api}, opts...)
}
