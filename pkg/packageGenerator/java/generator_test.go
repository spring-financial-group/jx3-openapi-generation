//go:build unit

package java_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that validates the configuration changes made to fix Java generation issues
func TestJavaGeneratorConfiguration(t *testing.T) {
	t.Run("Java OpenAPI configuration includes required properties", func(t *testing.T) {
		configPath := filepath.Join("..", "..", "..", "configs", "java-openapitools.json")
		assert.FileExists(t, configPath, "Java OpenAPI config should exist")

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var config map[string]interface{}
		err = json.Unmarshal(content, &config)
		require.NoError(t, err, "Config should be valid JSON")

		generatorCLI := config["generator-cli"].(map[string]interface{})
		generators := generatorCLI["generators"].(map[string]interface{})
		javaGen := generators["java"].(map[string]interface{})
		additionalProps := javaGen["additionalProperties"].(map[string]interface{})

		// These properties were added to fix the AbstractOpenApiSchema issues
		assert.Equal(t, "okhttp-gson", additionalProps["library"], 
			"Should use okhttp-gson library to fix generation issues")
		assert.Equal(t, "gson", additionalProps["serializationLibrary"], 
			"Should use gson serialization library")
	})

	t.Run("Java build template includes updated dependencies", func(t *testing.T) {
		templatePath := filepath.Join("..", "..", "..", "templates", "java", "build.gradle")
		content, err := os.ReadFile(templatePath)
		require.NoError(t, err)

		contentStr := string(content)

		// These dependency updates were made to fix compilation issues
		assert.Contains(t, contentStr, "io.swagger:swagger-annotations:1.6.14", 
			"Should use updated Swagger annotations")
		assert.Contains(t, contentStr, "com.squareup.okhttp3:okhttp:4.12.0", 
			"Should use updated OkHttp")
		assert.Contains(t, contentStr, "com.google.code.gson:gson:2.10.1", 
			"Should use updated Gson")
		assert.Contains(t, contentStr, "org.apache.oltu.oauth2:org.apache.oltu.oauth2.client:1.0.2", 
			"Should include OAuth2 client dependency to resolve import errors")
		assert.Contains(t, contentStr, "javax.annotation:javax.annotation-api:1.3.2", 
			"Should include javax annotations to resolve Generated annotation errors")
	})
}