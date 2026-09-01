// Package knife 将 Knife4j UI 与 OpenAPI 文档注册到 gin 路由组，
// 或任何实现了 Router 接口的框架适配器（见 gin/、huma/ 子包）。
package knife

import (
	"io"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"

	_gin "github.com/smarteng/knife4go/gin"
	_huma "github.com/smarteng/knife4go/huma"
)

// InitSwaggerKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 gin 路由组。
// 文档内容通过 Doc(doc) 显式提供（OpenAPI 3.0 JSON 字符串）；
// UI 页面路径默认 /doc.html，可用 DocPath(path) 修改。
// 注册位置（根级或路由组）不影响展示：前端按文档 paths 是否已带前缀自适应。
//
// 默认会临时抑制 gin 的 [GIN-debug] 路由日志，避免约 40 条 knife4j 静态资源
// 把用户业务路由的日志淹没；可通过 Verbose(true) 关闭该抑制。
func InitSwaggerKnife(router gin.IRouter, opts ...Opts) error {
	cfg := Config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	if !cfg.verbose && gin.IsDebugging() {
		defer withSilencedGinDebugOutput()()
	}

	return RegisterOpenAPI(&_gin.Router{R: router}, opts...)
}

// withSilencedGinDebugOutput 临时把 gin.DefaultWriter 重定向到 io.Discard，
// 返回一个用于恢复原 Writer 的函数；用法：defer withSilencedGinDebugOutput()()。
// 注意：非线程安全，仅用于服务启动阶段的路由注册（本身就应在单线程中完成）。
func withSilencedGinDebugOutput() func() {
	original := gin.DefaultWriter
	gin.DefaultWriter = io.Discard
	return func() {
		gin.DefaultWriter = original
	}
}

// InitHumaKnife 将 Knife4j UI、OpenAPI 文档端点与静态资产注册到 huma API。
// 文档内容通过 knife.Doc(doc) 显式提供（OpenAPI 3.0 JSON 字符串）；
// UI 页面路径默认 /doc.html，可用 knife.DocPath(path) 修改。
//
// 注册位置（根级 API 或 huma.NewGroup 前缀组）可自由选择：huma 组会把组前缀
// 写入文档 paths，前端检测到 paths 已带前缀时不再拼接 doc.html 位置前缀，
// 展示与调试请求直接使用文档自带的完整路径。
func InitHumaKnife(api huma.API, opts ...Opts) error {
	return RegisterOpenAPI(&_huma.Router{API: api}, opts...)
}
