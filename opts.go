package knife

// Config 是 InitSwaggerKnife / InitHumaKnife / RegisterOpenAPI 的注册配置。
type Config struct {
	docJson  string
	docPath  string
	basePath string
}

// Opts 以函数选项模式配置 knife4go 注册行为。
type Opts func(*Config)

// Doc 指定 OpenAPI 文档内容（JSON 字符串）。
// 未指定时默认读取 swag.ReadDoc("swagger")，需要空白导入 _ "项目/docs/swagger"。
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

// DocBasePath 声明 knife4go 注册位置的前缀（如 huma.NewGroup 创建的组前缀）。
// 注册在该前缀下时，若文档 paths 全部以该前缀开头，注册前会剥离此前缀：
// Knife4j 前端会基于 doc.html 所在路径自行拼接前缀，剥离后拼接结果恰好是
// 原始路径。注册在根级时无需声明。gin 入口会自动从路由组获取此前缀。
func DocBasePath(path string) Opts {
	return func(c *Config) {
		c.basePath = path
	}
}
