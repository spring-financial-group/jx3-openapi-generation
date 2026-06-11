package swagfilter

import (
	"encoding/json"
	"fmt"
)

var httpMethods = []string{"get", "post", "put", "delete", "patch", "head", "options"}

// StripTagFromSpec removes all occurrences of tagToStrip from every operation's
// tags array in the swagger JSON, leaving all other fields untouched.
func StripTagFromSpec(data []byte, tagToStrip string) ([]byte, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse swagger JSON: %w", err)
	}

	pathsRaw, ok := doc["paths"]
	if !ok {
		return json.MarshalIndent(doc, "", "  ")
	}

	var paths map[string]map[string]json.RawMessage
	if err := json.Unmarshal(pathsRaw, &paths); err != nil {
		return nil, fmt.Errorf("parse paths: %w", err)
	}

	for pathKey, pathItem := range paths {
		for _, method := range httpMethods {
			opRaw, ok := pathItem[method]
			if !ok {
				continue
			}
			var op map[string]json.RawMessage
			if err := json.Unmarshal(opRaw, &op); err != nil {
				continue
			}
			tagsRaw, ok := op["tags"]
			if !ok {
				continue
			}
			var tags []string
			if err := json.Unmarshal(tagsRaw, &tags); err != nil {
				continue
			}
			filtered := make([]string, 0, len(tags))
			for _, t := range tags {
				if t != tagToStrip {
					filtered = append(filtered, t)
				}
			}
			if newTagsRaw, err := json.Marshal(filtered); err == nil {
				op["tags"] = newTagsRaw
			}
			if newOpRaw, err := json.Marshal(op); err == nil {
				pathItem[method] = newOpRaw
			}
		}
		paths[pathKey] = pathItem
	}

	if newPathsRaw, err := json.Marshal(paths); err == nil {
		doc["paths"] = newPathsRaw
	}

	return json.MarshalIndent(doc, "", "  ")
}
