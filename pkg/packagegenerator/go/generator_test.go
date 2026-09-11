//go:build unit

package _go_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/stretchr/testify/require"
)

// Regression test for https://github.com/oapi-codegen/oapi-codegen/issues/697:
// a schema combining a single-element allOf with sibling properties must not
// silently drop those properties. This is the shape NSwag/System.Text.Json emit
// for polymorphic DTOs (base type via allOf, own fields as siblings), and it
// previously collapsed into a bare type alias to the base, silently losing the
// sibling fields. Fixed upstream in oapi-codegen v2.8.0 (PR #1415).
func TestGenerateCode_PreservesSiblingPropertiesOnSingleElementAllOf(t *testing.T) {
	const spec = `{
	  "openapi": "3.0.1",
	  "info": {"title": "t", "version": "1"},
	  "paths": {
	    "/derived": {
	      "get": {
	        "operationId": "getDerived",
	        "responses": {
	          "200": {
	            "description": "ok",
	            "content": {
	              "application/json": {"schema": {"$ref": "#/components/schemas/Derived"}}
	            }
	          }
	        }
	      }
	    }
	  },
	  "components": {
	    "schemas": {
	      "Base": {"type": "object", "properties": {"id": {"type": "string"}}},
	      "Derived": {
	        "type": "object",
	        "allOf": [{"$ref": "#/components/schemas/Base"}],
	        "properties": {"extra": {"type": "string"}},
	        "additionalProperties": false
	      }
	    }
	  }
	}`

	swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)

	code, err := codegen.Generate(swagger, codegen.Configuration{
		PackageName: "test",
		Generate:    codegen.GenerateOptions{Models: true},
	})
	require.NoError(t, err)
	require.NotContains(t, code, "type Derived = Base", "Derived must not collapse into a bare alias of Base")
	require.Contains(t, code, "Extra", "expected Derived struct to retain sibling property 'extra', got:\n"+code)
}
