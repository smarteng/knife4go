# knife4go

knife4go 是一个 Go 库，为 [gin](https://github.com/gin-gonic/gin) 服务挂载 [Knife4j](https://github.com/xiaoymin/knife4j) UI，把 swagger 文档渲染成开箱即用的接口文档页面。

**gin 与 swagger 一起使用**：knife4go 不解析接口注解、也不生成文档，它只负责把 OpenAPI 3.0 文档（OpenAPI JSON 字符串）注册到 gin，并提供 Knife4j UI 页面、文档端点与全部静态资产。文档通常由 [swag](https://github.com/swaggo/swag) 在你的项目中生成，经 `knife4go.Doc(doc)` 传入即可：

```
swag 生成文档 → swag.ReadDoc() → knife4go.Doc(doc) → InitSwaggerKnife(router) → Knife4j UI
```

一行接入，gin + swagger + Knife4j UI 一步到位。knife4go 自身不依赖 swag，swag 由宿主项目提供。

> 同时提供 `huma/` 子包，为 [Huma v2](https://huma.rocks/) 挂载同一套 UI；其他 Web 框架只需实现约 15 行的 `knife4go.Router` 接口即可接入（见「扩展新框架」）。

全部 UI 资产已数据化内嵌，运行时不依赖任何外部静态文件。

## 快速开始（gin）

```go
import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
	knife4go "github.com/jasonlabz/knife4go"

	_ "{module_path}/docs" // swag 生成的文档包，路径须与你项目的 swag 输出目录一致
)

// InitApiRouter 返回带 knife4go UI 的路由。
func InitApiRouter() *gin.Engine {
	router := gin.Default()
	// 注册到路由组时，端点自动带组前缀，如 /demo/doc.html
	serverGroup := router.Group("/demo")
	doc, err := swag.ReadDoc(swag.Name)
	if err != nil {
		panic(fmt.Errorf("read OpenAPI document from swag: %w", err))
	}
	if err := knife4go.InitSwaggerKnife(serverGroup, knife4go.Doc(doc)); err != nil {
		panic(err)
	}
	return router
}
```

流程：

1. 在宿主项目安装 swag（`go install github.com/swaggo/swag/cmd/swag@latest`），按 swag 规则在接口上写注解，运行 `swag init` 生成文档；
2. 空白导入生成的 `docs` 包，使文档在运行时完成注册；
3. 通过 `swag.ReadDoc(swag.Name)` 读取文档，经 `knife4go.Doc(doc)` 传入 `InitSwaggerKnife` 完成挂载。

### 依赖

- Go 1.25 及以上
- [gin](https://github.com/gin-gonic/gin) v1.12.0
- [swag](https://github.com/swaggo/swag)：knife4go 不依赖它，由宿主项目自行引入用于生成文档

最小可运行示例见 [`_examples/gin`](./_examples/gin)。

## huma（可选）

需要为 [Huma v2](https://huma.rocks/) 服务挂载 UI 时，使用 `huma/` 子包（此时才会引入 Huma v2 依赖）：

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

## 自定义文档与路径

`Doc()` 也可传入任意来源的 OpenAPI 3.0 JSON（不限于 swag）；UI 页面路径默认 `/doc.html`，可用 `DocPath()` 自定义：

```go
_ = knife4go.InitSwaggerKnife(router, knife4go.Doc(docJSON), knife4go.DocPath("/api-doc"))
```

## 注册位置

前端按文档 `paths` 是否已带注册位置前缀自适应，注册在根级还是带前缀的路由组都可以，无需声明前缀：

- **gin 路由组注册**（swag 文档无前缀）：前端基于 `doc.html` 位置拼接前缀，如注册在 `/demo` 组则 UI 位于 `/demo/doc.html`，调试请求自动带 `/demo` 前缀；
- **huma 前缀组注册**：`huma.NewGroup` 会把组前缀写入文档 `paths`，前端检测到已带前缀时不再拼接，直接使用文档中的完整路径；
- 文档始终**原样注册**，不做任何改写。

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

2. 调用 `knife4go.RegisterOpenAPI(router, opts...)`，或参照 `gin/`、`huma/` 子包封装 `InitSwaggerKnife` / `InitHumaKnife` 便捷入口。

## 示例效果

![img](./_examples/knife4go_example.png)

## Links

- [Knife4j](https://github.com/xiaoymin/knife4j)
- [Huma](https://huma.rocks/)

## 免责声明

本项目为公益项目，作者不对任何使用后果承担责任。

## License

[MIT LICENSE](./LICENSE)
