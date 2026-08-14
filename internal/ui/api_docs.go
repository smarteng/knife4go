package ui

import (
	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/knife4go/constant"
	"github.com/jasonlabz/knife4go/utils"
)

// API_DOCS_RELATIVE_PATH is the relative path that serves the OpenAPI document.
const API_DOCS_RELATIVE_PATH = constant.RootPath + "/v3/api-docs"

// AddOpenAPIDocumentRouter registers the serialized OpenAPI document endpoint.
func AddOpenAPIDocumentRouter(router gin.IRouter, path, document string) {
	if path == "" {
		path = API_DOCS_RELATIVE_PATH
	}
	utils.GetJson(router, path, document)
}
