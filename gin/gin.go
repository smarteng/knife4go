package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Router 将 gin.IRouter 适配为 Router。
type Router struct {
	R gin.IRouter
}

// GET 实现 Router.GET。
func (w *Router) GET(path, contentType string, content []byte) {
	w.R.GET(path, func(c *gin.Context) {
		c.Data(http.StatusOK, contentType, content)
	})
}

// BasePath 返回 gin 路由组前缀，RegisterOpenAPI 据此判断文档 paths 是否
// 已带注册位置前缀（gin.IRouter 接口未声明该方法，*Engine 与 *RouterGroup
// 均有，用类型断言获取）。
func (w *Router) BasePath() string {
	if routerWithBasePath, ok := w.R.(interface{ BasePath() string }); ok {
		return routerWithBasePath.BasePath()
	}
	return ""
}
