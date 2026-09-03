package knife

// Config 是 InitSwaggerKnife / InitHumaKnife / RegisterOpenAPI 的注册配置。
type Config struct {
	docJson     string
	docPath     string
	apiDocsPath string
	verbose     bool
	// template 前端模板名, 空串代表使用默认模板 ("ui")。
	// 可选值: "ui" (Vue2 旧版) / "vue" (Vue3 新版)；
	// 名称不在 templateRegistry 中时, RegisterOpenAPI 返回错误。
	template string

	// beforeStaticAssets 是可选的钩子，在批量注册 40 条静态资源前调用；
	// 返回一个用于收尾的清理函数（如恢复现场），在批量注册完成后立即执行。
	// 默认为 nil，此时不做任何包裹。适配器（如 gin）可通过此钩子实现日志静默等横切能力。
	beforeStaticAssets func() (cleanup func())
}

// Opts 以函数选项模式配置 knife4go 注册行为。
type Opts func(*Config)

// Doc 指定 OpenAPI 文档内容（OpenAPI 3.0 JSON 字符串）。
// knife4go 不读取任何文档生成器，注册时必须通过 Doc 传入文档。
func Doc(doc string) Opts {
	return func(c *Config) {
		c.docJson = doc
	}
}

// DocPath 指定 UI 页面路径，默认 /doc.html。
func DocPath(path string) Opts {
	return func(c *Config) {
		c.docPath = path
	}
}

// APIDocsPath 指定 OpenAPI 文档 JSON 端点路径，默认 /swagger/doc.json。
//
// 该路径同时会写入 swagger-config 响应体的 url 字段，供 Knife4j 前端加载文档。
// 传入空串等价于不设置（使用默认值）。示例：APIDocsPath("/api/openapi.json")。
func APIDocsPath(path string) Opts {
	return func(c *Config) {
		c.apiDocsPath = path
	}
}

// Verbose 控制 knife4go 注册静态资源时是否保留框架自身的路由日志。
//
// 默认（不传或 Verbose(false)）：InitSwaggerKnife 会在注册期间临时抑制
// gin 的 [GIN-debug] 路由输出，覆盖 40 条固定静态资源与 knife4j 前端约定的
// /v3/api-docs/swagger-config 探测端点，避免这些固定路由把用户业务路由日志淹没；
// UI 页面（DocPath）与文档 JSON（APIDocsPath）两条动态路由的日志始终保留。
// 注册结束后立即恢复，用户后续自行注册的路由日志不受影响。
// 传入 Verbose(true) 可关闭该抑制，恢复框架默认行为。
func Verbose(v bool) Opts {
	return func(c *Config) {
		c.verbose = v
	}
}

// Template 选择 knife4go 使用的前端模板。内置可选值: "ui" (Vue2 旧版) / "vue" (Vue3 新版)。
// 传入未登记的名称时, RegisterOpenAPI 返回包含可用名称列表的错误。
// 不调用时默认使用 "ui" 以保持向后兼容。
func Template(name string) Opts {
	return func(c *Config) {
		c.template = name
	}
}
