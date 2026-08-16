// Command huma 演示 knife4go 在 huma v2 框架下的最小接入方式。
package main

import (
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	knife "github.com/jasonlabz/knife4go"
)

func main() {
	router := gin.Default()
	api := humagin.New(router, huma.DefaultConfig("demo", "v1"))

	openAPIDocument, err := api.OpenAPI().Downgrade()
	if err != nil {
		panic(fmt.Errorf("downgrade huma OpenAPI document: %w", err))
	}
	_ = knife.InitHumaKnife(api, knife.Doc(string(openAPIDocument)))
	_ = router.Run(":8080")
}
