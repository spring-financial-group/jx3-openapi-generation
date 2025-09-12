//go:build unit

package java_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/openapitools"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/java"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// OpenAPI spec that triggers the specific issues we fixed
const testOpenAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {
    "/test": {
      "post": {
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/TestRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "OK"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "PurchasePrice": {
        "oneOf": [
          {"type": "number", "format": "double"},
          {"type": "string"}
        ]
      },
      "RequestedLoanAmount": {
        "oneOf": [
          {"type": "number", "format": "double"},
          {"type": "string"}
        ]
      },
      "TestRequest": {
        "type": "object",
        "properties": {
          "purchasePrice": {"$ref": "#/components/schemas/PurchasePrice"},
          "loanAmount": {"$ref": "#/components/schemas/RequestedLoanAmount"}
        }
      }
    }
  }
}`

// Test the actual generation behavior that was broken
func TestJavaGeneratorFixesIssues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping generation test in short mode")
	}

	// Create temporary directories
	tempDir := t.TempDir()
	specFile := filepath.Join(tempDir, "test-spec.json")
	outputDir := filepath.Join(tempDir, "generated")

	// Write the test spec that triggers oneOf/anyOf issues
	err := os.WriteFile(specFile, []byte(testOpenAPISpec), 0644)
	require.NoError(t, err)

	// Create a minimal config for testing
	cfg := &openapitools.Config{
		Schema: "test",
		Spaces: 2,
		GeneratorCLI: openapitools.GeneratorCLI{
			Version: "7.15.0",
			Generators: map[string]*openapitools.Generator{
				"java": {
					Name:      "java",
					InputSpec: specFile,
					Output:    outputDir,
					AdditionalProperties: map[string]string{
						"library":             "okhttp-gson",
						"serializationLibrary": "gson",
					},
				},
			},
		},
	}

	// Create base generator
	baseGen, err := packageGenerator.NewBaseGenerator(
		"1.0.0-test",
		"test-service", 
		"test-owner",
		"test-repo",
		"test-token",
		"test-user",
		specFile,
		"TestClient",
		cfg,
	)
	require.NoError(t, err)

	// Create Java generator
	javaGen := java.NewGenerator(baseGen)

	// Set the dynamic config variables that the Java generator normally sets
	cfg.GeneratorCLI.Generators["java"].AdditionalProperties["basePackage"] = "mqube.test-service" 
	cfg.GeneratorCLI.Generators["java"].AdditionalProperties["modelPackage"] = "mqube.test-service.models"

	// Generate the package - this is what was failing before
	// Note: We skip the template copying step that requires /templates/java since
	// that's only needed in the containerized environment for build.gradle templating
	generatedDir, err := baseGen.GeneratePackage(filepath.Join(outputDir, javaGen.GetPackageName()), domain.Java)
	require.NoError(t, err, "Java package generation should succeed")

	t.Run("AbstractOpenApiSchema is generated for oneOf schemas", func(t *testing.T) {
		// Check AbstractOpenApiSchema class exists (using the correct package structure)
		abstractSchemaPath := filepath.Join(generatedDir, "src", "main", "java", "mqube", "test_service", "models", "AbstractOpenApiSchema.java")
		assert.FileExists(t, abstractSchemaPath, "AbstractOpenApiSchema.java should be generated")

		content, err := os.ReadFile(abstractSchemaPath)
		require.NoError(t, err)
		contentStr := string(content)

		// Verify it has the methods that were missing
		assert.Contains(t, contentStr, "setActualInstance", "AbstractOpenApiSchema should have setActualInstance method")
		assert.Contains(t, contentStr, "getActualInstance", "AbstractOpenApiSchema should have getActualInstance method")
	})

	t.Run("oneOf schemas extend AbstractOpenApiSchema correctly", func(t *testing.T) {
		modelsDir := filepath.Join(generatedDir, "src", "main", "java", "mqube", "test_service", "models")

		// Check PurchasePrice model
		purchasePricePath := filepath.Join(modelsDir, "PurchasePrice.java")
		assert.FileExists(t, purchasePricePath, "PurchasePrice.java should be generated")

		content, err := os.ReadFile(purchasePricePath)
		require.NoError(t, err)
		contentStr := string(content)

		// These were the specific issues that caused compilation failures
		assert.Contains(t, contentStr, "extends AbstractOpenApiSchema", 
			"PurchasePrice should extend AbstractOpenApiSchema")
		assert.Contains(t, contentStr, "super.setActualInstance", 
			"PurchasePrice should call super.setActualInstance (this was failing)")
		assert.Contains(t, contentStr, "super.getActualInstance", 
			"PurchasePrice should call super.getActualInstance (this was failing)")

		// Check RequestedLoanAmount model  
		loanAmountPath := filepath.Join(modelsDir, "RequestedLoanAmount.java")
		assert.FileExists(t, loanAmountPath, "RequestedLoanAmount.java should be generated")

		content, err = os.ReadFile(loanAmountPath)
		require.NoError(t, err)
		contentStr = string(content)

		assert.Contains(t, contentStr, "extends AbstractOpenApiSchema", 
			"RequestedLoanAmount should extend AbstractOpenApiSchema")
		assert.Contains(t, contentStr, "super.setActualInstance", 
			"RequestedLoanAmount should call super.setActualInstance (this was failing)")
	})

	t.Run("generated code compiles with updated dependencies", func(t *testing.T) {
		// Copy our updated build.gradle template
		templatePath := filepath.Join("..", "..", "..", "templates", "java", "build.gradle")
		template, err := os.ReadFile(templatePath)
		require.NoError(t, err)

		// Replace template variables
		buildContent := string(template)
		buildContent = strings.ReplaceAll(buildContent, "{{ .GetPackageName }}", "com.test.generated")
		buildContent = strings.ReplaceAll(buildContent, "{{ .Version }}", "1.0.0-test")
		buildContent = strings.ReplaceAll(buildContent, "{{ .RepoName }}", "test-repo")
		buildContent = strings.ReplaceAll(buildContent, "{{ .GitUser }}", "test-user")
		buildContent = strings.ReplaceAll(buildContent, "{{ .GitToken }}", "test-token")

		// Remove publishing section for test
		lines := strings.Split(buildContent, "\n")
		var filteredLines []string
		inPublishing := false
		braceCount := 0

		for _, line := range lines {
			if strings.Contains(line, "publishing {") {
				inPublishing = true
				braceCount = 1
				continue
			}
			if inPublishing {
				braceCount += strings.Count(line, "{")
				braceCount -= strings.Count(line, "}")
				if braceCount == 0 {
					inPublishing = false
				}
				continue
			}
			filteredLines = append(filteredLines, line)
		}

		buildContent = strings.Join(filteredLines, "\n")
		
		// Write the build file
		buildPath := filepath.Join(generatedDir, "build.gradle")
		err = os.WriteFile(buildPath, []byte(buildContent), 0644)
		require.NoError(t, err)

		// Make gradlew executable
		gradlewPath := filepath.Join(generatedDir, "gradlew")
		err = os.Chmod(gradlewPath, 0755)
		require.NoError(t, err)

		// Attempt compilation - this was failing before our fixes
		cmd := exec.Command("./gradlew", "compileJava", "--no-daemon")
		cmd.Dir = generatedDir
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("Gradle output:\n%s", string(output))
			
			// Check for the specific errors that were happening before
			outputStr := string(output)
			if strings.Contains(outputStr, "cannot find symbol: class AbstractOpenApiSchema") {
				t.Fatal("Still getting AbstractOpenApiSchema symbol errors - the fix didn't work")
			}
			if strings.Contains(outputStr, "cannot find symbol.*super") {
				t.Fatal("Still getting super variable errors - the fix didn't work")
			}
			if strings.Contains(outputStr, "method does not override or implement a method from a supertype") {
				t.Fatal("Still getting @Override method signature errors - the fix didn't work")
			}
		}
		
		assert.NoError(t, err, "Java compilation should succeed with the fixes applied")
	})
}

// Basic test to validate our configuration is correct
func TestJavaGeneratorConfiguration(t *testing.T) {
	configPath := filepath.Join("..", "..", "..", "configs", "java-openapitools.json")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(content, &config)
	require.NoError(t, err)

	additionalProps := config["generator-cli"].(map[string]interface{})["generators"].(map[string]interface{})["java"].(map[string]interface{})["additionalProperties"].(map[string]interface{})

	assert.Equal(t, "okhttp-gson", additionalProps["library"])
	assert.Equal(t, "gson", additionalProps["serializationLibrary"])
}