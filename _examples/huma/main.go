// Command huma 演示 knife4go 在 huma v2 框架下的最小接入方式。
package main

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	knifehuma "github.com/jasonlabz/knife4go/huma"
)

func main() {
	router := gin.Default()
	api := humagin.New(router, huma.DefaultConfig("demo", "v1"))
	_ = knifehuma.InitSwaggerKnife(api)
	_ = router.Run(":8080")
}
