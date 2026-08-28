package js

import (
	"strings"
	"testing"
)

func TestAppSingleOpenAPISpecKeepsSchemasNamespace(t *testing.T) {
	content := string(App2fab4ac5Js.Content)

	if !strings.Contains(content, "c=this.readOpenApiSpeciOAS3(n,a),i.components={schemas:c}") {
		t.Fatal("expected the single-interface OpenAPI document to put filtered models under components.schemas")
	}
	if strings.Contains(content, "i.components=c") {
		t.Fatal("single-interface OpenAPI document must not put schemas directly under components")
	}
	if !strings.Contains(content, "#/components/schemas/") {
		t.Fatal("expected OpenAPI model references to keep the components.schemas namespace")
	}
}
