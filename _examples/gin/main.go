// Command gin 演示 knife4go 在 gin 框架下的最小接入方式。
package main

import (
	"github.com/gin-gonic/gin"
	knife4go "github.com/jasonlabz/knife4go"
)

// InitApiRouter 返回带 knife4go UI 的路由。
// 文档默认来自 swag.ReadDoc("swagger")（需空白导入 _ "项目/docs/swagger"），
// 也可用 knife4go.Doc(doc) 显式指定。
func InitApiRouter() *gin.Engine {
	router := gin.Default()
	serverGroup := router.Group("/demo")
	_ = knife4go.InitSwaggerKnife(serverGroup)
	return router
}

func main() {
	_ = InitApiRouter()
}
