//go:build unit

package swagfilter_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spring-financial-group/jx3-openapi-generation/pkg/swagfilter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripTagFromSpec(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		tag      string
		wantTags map[string][]string // "path:method" -> expected tags after strip
	}{
		{
			name: "StripsTargetTagFromOperation",
			input: map[string]interface{}{
				"paths": map[string]interface{}{
					"/api/cases": map[string]interface{}{
						"get": map[string]interface{}{
							"tags": []interface{}{"Cases", "external"},
						},
					},
				},
			},
			tag:      "external",
			wantTags: map[string][]string{"/api/cases:get": {"Cases"}},
		},
		{
			name: "LeavesOtherTagsUntouched",
			input: map[string]interface{}{
				"paths": map[string]interface{}{
					"/api/cases": map[string]interface{}{
						"get": map[string]interface{}{
							"tags": []interface{}{"Cases", "v2", "external"},
						},
					},
				},
			},
			tag:      "external",
			wantTags: map[string][]string{"/api/cases:get": {"Cases", "v2"}},
		},
		{
			name: "PreservesOperationWhenOnlyTagIsStripped",
			input: map[string]interface{}{
				"paths": map[string]interface{}{
					"/api/cases": map[string]interface{}{
						"get": map[string]interface{}{
							"tags": []interface{}{"external"},
						},
					},
				},
			},
			tag:      "external",
			wantTags: map[string][]string{"/api/cases:get": {}},
		},
		{
			name: "NoOpWhenTagNotPresent",
			input: map[string]interface{}{
				"paths": map[string]interface{}{
					"/api/cases": map[string]interface{}{
						"get": map[string]interface{}{
							"tags": []interface{}{"Cases"},
						},
					},
				},
			},
			tag:      "external",
			wantTags: map[string][]string{"/api/cases:get": {"Cases"}},
		},
		{
			name: "HandlesMultiplePathsAndMethods",
			input: map[string]interface{}{
				"paths": map[string]interface{}{
					"/api/cases": map[string]interface{}{
						"get":  map[string]interface{}{"tags": []interface{}{"Cases", "external"}},
						"post": map[string]interface{}{"tags": []interface{}{"Cases"}},
					},
					"/api/events": map[string]interface{}{
						"get": map[string]interface{}{"tags": []interface{}{"Events", "external"}},
					},
				},
			},
			tag: "external",
			wantTags: map[string][]string{
				"/api/cases:get":  {"Cases"},
				"/api/cases:post": {"Cases"},
				"/api/events:get": {"Events"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			require.NoError(t, err)

			result, err := swagfilter.StripTagFromSpec(data, tc.tag)
			require.NoError(t, err)

			var doc map[string]interface{}
			require.NoError(t, json.Unmarshal(result, &doc))

			paths := doc["paths"].(map[string]interface{})
			for key, wantTags := range tc.wantTags {
				colonIdx := strings.LastIndex(key, ":")
				pathStr := key[:colonIdx]
				method := key[colonIdx+1:]

				pathItem := paths[pathStr].(map[string]interface{})
				op := pathItem[method].(map[string]interface{})
				var gotTags []string
				if rawTags, ok := op["tags"].([]interface{}); ok {
					gotTags = make([]string, len(rawTags))
					for i, t := range rawTags {
						gotTags[i] = t.(string)
					}
				}
				assert.ElementsMatch(t, wantTags, gotTags, "tags for %s %s", method, pathStr)
			}
		})
	}
}

func TestStripTagFromSpec_PreservesNonPathsFields(t *testing.T) {
	input := map[string]interface{}{
		"swagger": "2.0",
		"info":    map[string]interface{}{"title": "My API", "version": "v1"},
		"definitions": map[string]interface{}{
			"MyDTO": map[string]interface{}{"type": "object"},
		},
		"paths": map[string]interface{}{
			"/api/cases": map[string]interface{}{
				"get": map[string]interface{}{"tags": []interface{}{"Cases", "external"}},
			},
		},
	}
	data, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := swagfilter.StripTagFromSpec(data, "external")
	require.NoError(t, err)

	var doc map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &doc))

	assert.Equal(t, "2.0", doc["swagger"])
	info := doc["info"].(map[string]interface{})
	assert.Equal(t, "My API", info["title"])
	defs := doc["definitions"].(map[string]interface{})
	assert.Contains(t, defs, "MyDTO")
}

func TestStripTagFromSpec_InvalidJSON(t *testing.T) {
	_, err := swagfilter.StripTagFromSpec([]byte("not valid json"), "external")
	assert.Error(t, err)
}

func TestStripTagFromSpec_NoPathsField(t *testing.T) {
	input := map[string]interface{}{
		"swagger": "2.0",
		"info":    map[string]interface{}{"title": "test"},
	}
	data, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := swagfilter.StripTagFromSpec(data, "external")
	require.NoError(t, err)

	var doc map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &doc))
	assert.Equal(t, "2.0", doc["swagger"])
}

func TestStripTagFromSpec_MaestroFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/maestro.json")
	require.NoError(t, err)

	result, err := swagfilter.StripTagFromSpec(data, "external")
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/maestro_filtered.json")
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(result))
}
