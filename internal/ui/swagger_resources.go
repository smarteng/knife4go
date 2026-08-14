package ui

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/knife4go/constant"
	"github.com/jasonlabz/knife4go/utils"
)

// OPENAPI_CONFIG_PATH is the relative path that serves the embedded UI configuration.
const OPENAPI_CONFIG_PATH = constant.RootPath + "/v3/api-docs/swagger-config"

// AddOpenAPIConfigRouter registers the embedded UI configuration for documentURL.
func AddOpenAPIConfigRouter(router gin.IRouter, path, documentURL string) {
	if path == "" {
		path = OPENAPI_CONFIG_PATH
	}

	basePath := strings.TrimSuffix(documentURL, API_DOCS_RELATIVE_PATH)
	content := fmt.Sprintf(`{"configUrl": %q,"oauth2RedirectUrl": %q,"url": %q,"validatorUrl": ""}`,
		basePath+OPENAPI_CONFIG_PATH,
		basePath+constant.RootPath+"/swagger-ui/oauth2-redirect.html",
		documentURL,
	)
	utils.GetJson(router, path, content)
}
