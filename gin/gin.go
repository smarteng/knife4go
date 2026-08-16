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
