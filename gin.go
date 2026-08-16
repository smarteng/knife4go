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
