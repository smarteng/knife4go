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
// 注册位置可自由选择：注册在 huma.NewGroup 前缀组上时，用 DocBasePath 声明
// 组前缀，knife4go 会相应剥离文档 paths 中的该前缀（Knife4j 前端会基于
// doc.html 所在路径自行拼接前缀，剥离后拼接结果恰好是原始路径）；注册在
// 根级 API 时无需声明。
func InitHumaKnife(api huma.API, opts ...Opts) error {
	return RegisterOpenAPI(&_huma.Router{API: api}, opts...)
}
