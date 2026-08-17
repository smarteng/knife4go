package knife

// Config 是 InitSwaggerKnife / InitHumaKnife / RegisterOpenAPI 的注册配置。
type Config struct {
	docJson string
	docPath string
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
