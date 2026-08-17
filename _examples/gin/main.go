// Command gin 演示 knife4go 在 gin 框架下的最小接入方式。
package main

import (
	"github.com/gin-gonic/gin"
	knife4go "github.com/jasonlabz/knife4go"
)

// openAPI303Document 是最小 OpenAPI 3.0 文档示例。
// 实际项目可从 swag 等生成器读取后经 knife4go.Doc 传入。
const openAPI303Document = `{"openapi":"3.0.3","info":{"title":"demo","version":"v1"},"paths":{}}`

// InitApiRouter 返回带 knife4go UI 的路由。
func InitApiRouter() *gin.Engine {
	router := gin.Default()
	serverGroup := router.Group("/demo")
	_ = knife4go.InitSwaggerKnife(serverGroup, knife4go.Doc(openAPI303Document))
	return router
}

func main() {
	_ = InitApiRouter()
}
