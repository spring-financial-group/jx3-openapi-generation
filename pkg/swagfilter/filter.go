package swagfilter

import (
	"encoding/json"
	"fmt"

	"github.com/go-openapi/spec"
)

// StripTagFromSpec removes all occurrences of tagToStrip from every operation's
// tags array in the swagger JSON, leaving all other fields untouched.
func StripTagFromSpec(data []byte, tagToStrip string) ([]byte, error) {
	var swagger spec.Swagger
	if err := swagger.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("parse swagger JSON: %w", err)
	}

	if swagger.Paths != nil {
		for _, pathItem := range swagger.Paths.Paths {
			for _, op := range pathItemOperations(pathItem) {
				if op == nil {
					continue
				}
				filtered := make([]string, 0, len(op.Tags))
				for _, t := range op.Tags {
					if t != tagToStrip {
						filtered = append(filtered, t)
					}
				}
				op.Tags = filtered
			}
		}
	}

	return json.MarshalIndent(swagger, "", "  ")
}

func pathItemOperations(item spec.PathItem) []*spec.Operation {
	return []*spec.Operation{
		item.Get, item.Post, item.Put, item.Delete,
		item.Patch, item.Head, item.Options,
	}
}
