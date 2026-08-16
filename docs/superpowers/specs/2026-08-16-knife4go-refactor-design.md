# knife4go 重构设计：框架无关核心 + gin/huma 适配器

日期：2026-08-16

## 背景与目标

knife4go 当前 HEAD 引入了 `Init(router gin.IRouter, provider DocumentProvider, ...)` 新 API 和 `internal/` 目录结构。而生产项目 dagine、dagine-dashboard 引用的是 router.go:45 对应的旧版本（pseudo-version `v1.0.1-0.20241118142759-6386e3973279`，即提交 6386e39），API 为 `InitSwaggerKnife(router gin.IRouter, opts ...Opts)`，默认通过 `swag.ReadDoc("swagger")` 读取 OpenAPI 文档，选项为 `Doc` / `DocPath`。

重构目标（用户确认）：

1. **回退 API 形态**：恢复 `InitSwaggerKnife` + `Doc`/`DocPath` + `swag.ReadDoc` 默认取文档；删除 `DocumentProvider` / `Init` 新 API。
2. **支持 gin 与 huma，并易于扩展其他 web/api 框架**：核心逻辑框架无关，gin 与 huma 各一个薄适配器；新框架只需实现一个约 15 行的接口。
3. **目录结构调整，去掉 internal 层**：资产目录公开化，删除 `constant/`、`utils/`。

## 架构

```
knife4go/
├── go.mod            gin + swaggo/swag 直接依赖；huma 仅 huma/ 子包引入
├── knife.go          InitSwaggerKnife(gin.IRouter, ...Opts) 兼容入口 + 文档解析（swag.ReadDoc 默认）
├── gin.go            gin 适配器：ginRouter 实现 Router 接口
├── register.go       Router 接口 + 核心注册流程 Register(Router, opts)（框架无关）
├── opts.go           Opts / Doc / DocPath 选项
├── ui/               资产数据（原 internal/ui/knife 迁出并公开）
│   ├── assets.go     Asset 类型 + 全部资产汇总切片
│   ├── doc_html.go   doc.html 页面
│   ├── icons/        favicon 等图标资产
│   └── webjars/      css / fonts / img / js / oauth 资产
├── huma/             huma 适配器：InitSwaggerKnife(huma.API, ...Opts)
├── knife_test.go / gin_test.go / register_test.go   核心与 gin 测试
├── huma/huma_test.go                                  huma 适配器测试
└── _examples/        gin 与 huma 用法示例
```

删除内容：`internal/` 整体、`constant/`（RootPath 为空串，内联进核心）、`utils/`（gin 实现移入 gin.go）。

## 核心接口（框架扩展点）

```go
// register.go
// Router 是 knife4go 注册静态 GET 路由所需的最小框架无关接口。
// 新增框架支持：实现该接口后调用 Register，或参照 huma/ 提供便捷入口包。
type Router interface {
    // GET 注册一条 GET 路由，响应体固定为 content，Content-Type 为 contentType。
    GET(path, contentType string, content []byte)
}
```

**gin 适配**（根包 gin.go，约 15 行）：

```go
type ginRouter struct{ r gin.IRouter }

func (w *ginRouter) GET(path, contentType string, content []byte) {
    w.r.GET(path, func(c *gin.Context) {
        c.Data(http.StatusOK, contentType, content)
    })
}
```

**huma 适配**（huma/ 包，约 20 行）：已验证 huma v2 的 `API.Handle(op *Operation, handler func(ctx Context))` 支持裸 handler（huma 自身用它提供 spec JSON），`Operation.Hidden: true` 保证静态路由不污染 huma 生成的 OpenAPI 文档：

```go
type humaRouter struct{ api huma.API }

func (w *humaRouter) GET(path, contentType string, content []byte) {
    w.api.Handle(&huma.Operation{
        Method: http.MethodGet,
        Path:   path,
        Hidden: true,
    }, func(ctx huma.Context) {
        ctx.SetStatus(http.StatusOK)
        ctx.SetHeader("Content-Type", contentType)
        _, _ = ctx.BodyWriter().Write(content)
    })
}
```

## 核心注册流程

`Register(router Router, opts ...Opts) error` 职责：

1. 解析 OpenAPI 文档：`Doc(doc)` 提供时用之；否则 `swag.ReadDoc("swagger")`。解析失败返回 error。
2. 构建并注册动态路由（内容依赖文档与 basePath，运行时计算）：
   - `/doc.html`（`DocPath` 可配，沿用 knife 的 `AddRouterOfDocHtml` 逻辑）
   - `/v3/api-docs`（OpenAPI JSON）
   - `/v3/api-docs/swagger-config`（含 configUrl / oauth2RedirectUrl / url / validatorUrl 的 UI 配置，计算逻辑与现状一致）
3. 注册全部静态资产：遍历 `ui` 包汇总的资产切片，逐个 `router.GET(asset.Path, asset.ContentType, asset.Content)`。

`InitSwaggerKnife(router gin.IRouter, opts ...Opts) error` 与 `huma.InitSwaggerKnife(api huma.API, opts ...Opts) error` 均为 `Register(适配器{...}, opts...)` 的一行包装，保持与 6386e39 完全一致的调用形态与行为。

## 资产数据化

约 150 个前端资产文件（internal/ui/knife 下的 webjars、icons）从 `AddRouterOfXxx(router gin.IRouter, path)` 函数改写为纯数据声明：

```go
// ui/webjars/css/assets.go 示例
var CSSAssets = []Asset{
    {Path: "/webjars/css/app.ac23e017.css", ContentType: "text/css", Content: []byte(`...`)},
    // ...
}
```

```go
// ui/assets.go
type Asset struct {
    Path        string
    ContentType string
    Content     []byte
}

// AllAssets 返回全部静态资产（icons + webjars）。
func AllAssets() []Asset { ... }
```

- 原 hex 编码的资产（GetJs / GetOther 对应的 js/其他二进制）在包内用一个私有 helper 在声明期解码为 `Content []byte`。
- 原 `log.Fatal` 写错误处理删除；注册失败由框架自身行为决定（如 gin 路由冲突 panic），核心不做额外包装。

## 测试

- **核心（register_test.go）**：用记录型假 Router 收集注册调用，断言全部资产路径、content-type、内容非空；断言动态路由（doc.html、api-docs、swagger-config）注册在 `DocPath` 与 `Doc` 选项下行为正确。资产路径断言覆盖 6386e39 的全部路由，保证回归。
- **gin（gin_test.go / knife_test.go）**：gin.New + 适配器注册后，httptest 请求关键路径（doc.html、api-docs、swagger-config、若干 webjars 资产）断言状态码与响应体。
- **huma（huma/huma_test.go）**：`humatest.NewRouter` + `InitSwaggerKnife` 注册后，请求关键路径断言状态码与响应体；并断言 `OpenAPI()` 输出中不包含 knife4go 静态路由（Hidden 生效）。

## 外部调用点迁移

| 项目 | 现状 | 迁移 |
|------|------|------|
| dagine / dagine-dashboard | `knife4go.InitSwaggerKnife(serverGroup)`（pin 6386e39） | **零改动**（API 与行为一致） |
| generate-example-project | `knife.Init(serverGroup, knife.DocumentProviderFunc(huma 文档))`（pin HEAD） | 改为 `knife4go.InitSwaggerKnife(serverGroup, knife4go.Doc(huma 文档字符串))` |

## 其他

- go.mod：恢复 `swaggo/swag` 依赖；gin 保持直接依赖（根包入口需要）；huma 只出现在 huma/ 子包的 import 中。
- README：更新为 gin + huma 双框架用法与扩展指南。
- 提交后由用户决定版本 tag 与推送。
