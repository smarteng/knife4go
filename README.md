# knife4go

Knife4j UI for caller-provided OpenAPI 3.0 JSON documents served with Gin.

Requires Go 1.21 or later.

## Usage

Callers supply an already-generated OpenAPI 3.0 JSON document. Huma is a
compatible document producer, but it is not a Knife4go dependency.

```go
import (
	"os"

	"github.com/gin-gonic/gin"
	knife4go "github.com/jasonlabz/knife4go"
)

func InitRouters() *gin.Engine {
	router := gin.Default()
	provider := knife4go.DocumentProviderFunc(func() ([]byte, error) {
		return os.ReadFile("openapi.json")
	})
	if err := knife4go.Init(router, provider); err != nil {
		panic(err)
	}
	return router
}
```

Open the UI at `/doc.html`. The OpenAPI JSON remains available at `/v3/api-docs`.
## Disclaimer
Public welfare projects.  
The disclaimer asserts that the individual won't be held responsible for any.

## MIT LICENSE
[LICENSE](./LICENSE)

# Links

- [Knife4j](https://github.com/xiaoymin/knife4j)
- [Huma](https://huma.rocks/)
