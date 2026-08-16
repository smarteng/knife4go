// Package knife 将 Knife4j UI 与 OpenAPI 文档注册到 gin 路由组，
// 或任何实现了 Router 接口的框架适配器（见 gin/、huma/ 子包）。
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
// 注册位置（根级或路由组）不影响展示：前端按文档 paths 是否已带前缀自适应。
func InitSwaggerKnife(router gin.IRouter, opts ...Opts) error {
	return RegisterOpenAPI(&_gin.Router{R: router}, opts...)
}

// InitHumaKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 huma API。
// 文档来源默认 swag.ReadDoc("swagger")，也可通过 knife.Doc(doc) 显式指定；
// UI 页面路径默认 /doc.html，可用 knife.DocPath(path) 修改。
//
// 注册位置（根级 API 或 huma.NewGroup 前缀组）可自由选择：huma 组会把组前缀
// 写入文档 paths，前端检测到 paths 已带前缀时不再拼接 doc.html 位置前缀，
// 展示与调试请求直接使用文档自带的完整路径。
func InitHumaKnife(api huma.API, opts ...Opts) error {
	return RegisterOpenAPI(&_huma.Router{API: api}, opts...)
}
