# knife4go

knife4go 是一个 Go 库，为 Web 服务挂载 [Knife4j](https://github.com/xiaoymin/knife4j) UI 与 OpenAPI 3.0 文档：提供 UI 页面、OpenAPI 文档端点、UI 配置端点及全部静态资产。采用"框架无关核心 + 适配器"架构：

- 根包适配 [gin](https://github.com/gin-gonic/gin)，入口 `InitSwaggerKnife`，与旧版用法一致；
- `huma/` 子包适配 [Huma v2](https://huma.rocks/)，入口同为 `InitSwaggerKnife`；
- 其他 Web/API 框架只需实现约 15 行的 `knife4go.Router` 接口即可接入（见「扩展新框架」）。

全部 UI 资产已数据化内嵌，运行时不依赖任何外部静态文件。

## 依赖要求

- Go 1.25 及以上
- [gin](https://github.com/gin-gonic/gin) v1.12.0
- [Huma v2](https://github.com/danielgtaylor/huma/v2)（仅使用 `huma/` 子包时需要）
- [swaggo/swag](https://github.com/swaggo/swag)（文档默认来源，可用 `Doc()` 替代）

## 快速开始

### gin

```go
import (
	"github.com/gin-gonic/gin"
	knife4go "github.com/jasonlabz/knife4go"
)

// InitApiRouter 返回带 knife4go UI 的路由。
func InitApiRouter() *gin.Engine {
	router := gin.Default()
	// 注册到路由组时，端点自动带组前缀，如 /demo/doc.html
	serverGroup := router.Group("/demo")
	if err := knife4go.InitSwaggerKnife(serverGroup); err != nil {
		panic(err)
	}
	return router
}
```

最小可运行示例见 [`_examples/gin`](./_examples/gin)。

### huma

```go
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
```

最小可运行示例见 [`_examples/huma`](./_examples/huma)。

**前端适配说明**：内嵌 UI 资产保持与既有版本一致，仅在 `app.js` 上打字节级
补丁：springdoc 分支
尊重 `appendBasePath` 检测——文档 `paths` 已带注册前缀（huma 组场景）时不再拼接
`doc.html` 位置前缀。

**注册位置自由选择**：knife4j 前端检测文档 `paths` 是否已带注册位置前缀（`huma.NewGroup` 会把组前缀写入文档 `paths`），已带则不拼接 `doc.html` 位置前缀，未带则按原逻辑拼接。因此：

- **huma 前缀组注册**：文档 `paths` 自带组前缀（用 `serverAPI.OpenAPI()` 生成文档），前端不再拼接，展示与调试请求直接使用文档中的完整路径；
- **gin 路由组注册**（swag 文档无前缀）：前端按原逻辑拼接 `doc.html` 位置前缀，行为与旧版一致；
- 文档始终**原样注册**，不做任何改写，也无需声明前缀。

## 文档来源

OpenAPI 文档来源有两种：

1. 默认读取 [swag](https://github.com/swaggo/swag) 生成文档：`swag.ReadDoc("swagger")`，需空白导入生成的 docs 包：

   ```go
   _ "你的项目/docs/swagger"
   ```

2. 用 `Doc()` 显式指定文档内容（OpenAPI 3.0 JSON 字符串）：

   ```go
   _ = knife4go.InitSwaggerKnife(router, knife4go.Doc(docJSON))
   ```

UI 页面路径默认 `/doc.html`，可用 `DocPath()` 自定义：

```go
_ = knife4go.InitSwaggerKnife(router, knife4go.Doc(docJSON), knife4go.DocPath("/api-doc"))
```

## HTTP 端点

注册后提供以下端点（注册在带前缀的路由组时自动带组前缀，如 `/demo/doc.html`）：

| 路径 | 说明 |
| --- | --- |
| `/doc.html` | Knife4j UI 页面（路径可用 `DocPath()` 修改） |
| `/v3/api-docs` | OpenAPI 3.0 文档（JSON） |
| `/v3/api-docs/swagger-config` | UI 配置端点（Knife4j 前端据此加载文档） |

## 扩展新框架

knife4go 的注册逻辑是框架无关的，接入新框架只需两步：

1. 实现 `knife4go.Router` 接口（仅一个方法，约 15 行），把每条静态路由注册为返回固定内容的 GET：

   ```go
   type Router interface {
   	// GET 注册一条 GET 路由，响应体固定为 content，Content-Type 为 contentType（空串表示不设置）。
   	GET(path, contentType string, content []byte)
   }
   ```

2. 调用 `knife4go.Register(router, opts...)`，或参照 `huma/` 子包封装 `InitSwaggerKnife` 便捷入口。

## 示例效果

![img](./_examples/knife4go_example.png)

## Links

- [Knife4j](https://github.com/xiaoymin/knife4j)
- [Huma](https://huma.rocks/)

## 免责声明

本项目为公益项目，作者不对任何使用后果承担责任。

## License

[MIT LICENSE](./LICENSE)
