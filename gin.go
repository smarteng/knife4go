package knife

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginRouter 将 gin.IRouter 适配为 Router。
type ginRouter struct {
	r gin.IRouter
}

// GET 实现 Router.GET。
func (w *ginRouter) GET(path, contentType string, content []byte) {
	w.r.GET(path, func(c *gin.Context) {
		c.Data(http.StatusOK, contentType, content)
	})
}

// BasePath 返回路由组前缀，Register 据此计算 OpenAPI 文档的完整 URL。
func (w *ginRouter) BasePath() string {
	// gin.IRouter 接口未暴露 BasePath，*Engine 与 *RouterGroup 均有该方法，用类型断言取前缀
	if routerWithBasePath, ok := w.r.(interface{ BasePath() string }); ok {
		return routerWithBasePath.BasePath()
	}
	return ""
}
