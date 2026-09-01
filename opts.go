package knife

// Config 是 InitSwaggerKnife / InitHumaKnife / RegisterOpenAPI 的注册配置。
type Config struct {
	docJson string
	docPath string
	verbose bool
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

// Verbose 控制 knife4go 注册静态资源时是否保留框架自身的路由日志。
//
// 默认（不传或 Verbose(false)）：InitSwaggerKnife 会在注册期间临时抑制
// gin 的 [GIN-debug] 路由输出，避免 40 条固定的静态资源路由把业务路由日志淹没；
// 注册结束后立即恢复，用户后续自行注册的路由日志不受影响。
// 传入 Verbose(true) 可关闭该抑制，恢复框架默认行为。
func Verbose(v bool) Opts {
	return func(c *Config) {
		c.verbose = v
	}
}
