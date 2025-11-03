package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/cmd/generate"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/file"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/rootcmd"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"
)

// Options for the test command
type Options struct {
	Languages []string
	SpecPath  string
	FileIO    domain.FileIO
}

var (
	testLong = templates.LongDesc(`
		Test package generation by dogfooding the CLI with sensible defaults.

		This command automatically:
		- Detects or creates a test swagger specification
		- Sets up test environment variables with sensible defaults
		- Generates packages for specified languages
		- Validates the generated packages
	`)

	testExample = templates.Examples(`
		# Test all supported languages
		%s test

		# Test specific languages
		%s test csharp java

		# Test with a custom swagger spec
		%s test --spec-path ./my-swagger.json go python
	`)
)

const defaultTestSwagger = `{
  "swagger": "2.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0",
    "description": "Test API for validating OpenAPI package generation"
  },
  "host": "api.example.com",
  "basePath": "/v1",
  "schemes": ["https"],
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check endpoint",
        "operationId": "getHealth",
        "responses": {
          "200": {
            "description": "Service is healthy",
            "schema": {
              "$ref": "#/definitions/HealthResponse"
            }
          }
        }
      }
    },
    "/users/{userId}": {
      "get": {
        "summary": "Get user by ID",
        "operationId": "getUserById",
        "parameters": [
          {
            "name": "userId",
            "in": "path",
            "required": true,
            "type": "string",
            "description": "User ID"
          }
        ],
        "responses": {
          "200": {
            "description": "User found",
            "schema": {
              "$ref": "#/definitions/User"
            }
          },
          "404": {
            "description": "User not found"
          }
        }
      }
    }
  },
  "definitions": {
    "HealthResponse": {
      "type": "object",
      "properties": {
        "status": {
          "type": "string"
        },
        "timestamp": {
          "type": "string",
          "format": "date-time"
        }
      }
    },
    "User": {
      "type": "object",
      "required": ["id", "email"],
      "properties": {
        "id": {
          "type": "string"
        },
        "email": {
          "type": "string",
          "format": "email"
        },
        "name": {
          "type": "string"
        },
        "createdAt": {
          "type": "string",
          "format": "date-time"
        }
      }
    }
  }
}`

// NewCmdTest creates a command object for testing package generation
func NewCmdTest() *cobra.Command {
	o := &Options{
		FileIO: file.NewFileIO(),
	}

	cmd := &cobra.Command{
		Use:     "test [languages...]",
		Short:   "Test package generation with sensible defaults",
		Long:    testLong,
		Example: fmt.Sprintf(testExample, rootcmd.BinaryName, rootcmd.BinaryName, rootcmd.BinaryName),
		// Don't validate on creation - we'll set env vars first
		DisableFlagParsing: false,
		Run: func(cmd *cobra.Command, args []string) {
			o.Languages = args
			err := o.Run()
			helper.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&o.SpecPath, "spec-path", "s", "", "Path to custom swagger specification (optional)")

	return cmd
}

// Run executes the test command
func (o *Options) Run() error {
	log.Info().Msg("🔧 OpenAPI Package Generation Test")
	log.Info().Msg("==========================================")

	// Setup test environment
	if err := o.setupTestEnvironment(); err != nil {
		return err
	}

	// Setup swagger spec
	specPath, err := o.ensureSwaggerSpec()
	if err != nil {
		return err
	}

	// Set environment variables for testing
	if err := o.setTestEnvironmentVariables(specPath); err != nil {
		return err
	}

	// Determine languages to test
	languages := o.getLanguages()
	log.Info().Msgf("📦 Testing languages: %v", languages)

	// Run the generate packages command
	log.Info().Msg("")
	log.Info().Msg("🚀 Generating packages with build validation...")

	// Now that environment variables are set, create the generate command
	// This will initialize with the environment variables we just set
	generateCmd := generate.NewCmdGenerate()

	// Create the full command line args (generate packages <languages>)
	args := append([]string{"packages"}, languages...)
	generateCmd.SetArgs(args)

	if err := generateCmd.Execute(); err != nil {
		log.Error().Msgf("❌ Package generation failed: %v", err)
		return err
	}

	log.Info().Msg("")
	log.Info().Msg("✅ All packages generated successfully!")
	log.Info().Msg("🎉 Package generation test completed!")

	return nil
}

func (o *Options) setupTestEnvironment() error {
	// Configure git if available
	log.Info().Msg("🔧 Configuring git for testing...")
	os.Setenv("GIT_AUTHOR_NAME", getEnvOrDefault("GIT_AUTHOR_NAME", "test-bot"))
	os.Setenv("GIT_AUTHOR_EMAIL", getEnvOrDefault("GIT_AUTHOR_EMAIL", "test-bot@example.com"))

	return nil
}

func (o *Options) getLanguages() []string {
	if len(o.Languages) > 0 {
		return o.Languages
	}

	// Default to all supported languages except go (which requires git operations)
	return []string{"angular", "csharp", "java", "typescript", "python", "rust"}
}

func (o *Options) ensureSwaggerSpec() (string, error) {
	// If user provided a spec path, use it
	if o.SpecPath != "" {
		absPath, err := filepath.Abs(o.SpecPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve spec path: %w", err)
		}

		exists, err := o.FileIO.Exists(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to check if spec exists: %w", err)
		}
		if !exists {
			return "", fmt.Errorf("spec file not found at %s", absPath)
		}

		log.Info().Msgf("📋 Using custom swagger spec: %s", absPath)
		return absPath, nil
	}

	// Try to find swagger.json in common locations
	searchPaths := []string{
		"./mocks/swagger.json",
		"./swagger.json",
		"./test/swagger.json",
		"./spec/swagger.json",
	}

	for _, path := range searchPaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		exists, err := o.FileIO.Exists(absPath)
		if err != nil {
			continue
		}
		if exists {
			log.Info().Msgf("📋 Found swagger spec at: %s", absPath)
			return absPath, nil
		}
	}

	// No spec found, create a default one
	log.Info().Msg("📋 No swagger spec found, creating default test spec...")

	// Ensure mocks directory exists
	mocksDir := "./mocks"
	if err := os.MkdirAll(mocksDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create mocks directory: %w", err)
	}

	specPath := filepath.Join(mocksDir, "swagger.json")
	if err := o.FileIO.Write(specPath, []byte(defaultTestSwagger), 0644); err != nil {
		return "", fmt.Errorf("failed to write test swagger spec: %w", err)
	}

	absPath, err := filepath.Abs(specPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve spec path: %w", err)
	}

	log.Info().Msgf("📋 Created test swagger spec at: %s", absPath)
	return absPath, nil
}

func (o *Options) setTestEnvironmentVariables(specPath string) error {
	log.Info().Msg("🔧 Setting test environment variables...")

	// Set defaults for all required environment variables
	// Note: Some generators (e.g., Python) use these values to construct branch names and paths
	// so we need to ensure they form valid identifiers
	envVars := map[string]string{
		"SwaggerServiceName": getEnvOrDefault("SwaggerServiceName", "test-service"),
		"SpecPath":           specPath,
		"VERSION":            getEnvOrDefault("VERSION", "0.0.0-test"),
		"REPO_OWNER":         getEnvOrDefault("REPO_OWNER", "test-owner"),
		"REPO_NAME":          getEnvOrDefault("REPO_NAME", "test-repo"),
		"PackageName":        getEnvOrDefault("PackageName", "TestClient"),
		"GIT_USER":           getEnvOrDefault("GIT_USER", "test-user"),
		"GIT_TOKEN":          getEnvOrDefault("GIT_TOKEN", "test-token"),
		"SKIP_PUSH":          getEnvOrDefault("SKIP_PUSH", "true"),
	}

	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Log the configuration
	log.Info().Msg("🔄 Test configuration:")
	log.Info().Msgf("  SwaggerServiceName: %s", envVars["SwaggerServiceName"])
	log.Info().Msgf("  SpecPath: %s", envVars["SpecPath"])
	log.Info().Msgf("  VERSION: %s", envVars["VERSION"])
	log.Info().Msgf("  REPO_OWNER: %s", envVars["REPO_OWNER"])
	log.Info().Msgf("  REPO_NAME: %s", envVars["REPO_NAME"])

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
