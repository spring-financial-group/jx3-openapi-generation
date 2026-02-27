# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go CLI tool that wraps the OpenAPI Generator to create client packages in multiple languages from OpenAPI specifications. It's designed to work with Jenkins X pipelines and automatically pushes generated packages to repositories.

## Development Commands

### Build
- `make build` - Build the binary for current OS
- `make build-all` - Build all files including tests
- `make linux` - Build for Linux
- `make darwin` - Build for macOS
- `make win` - Build for Windows

### Testing
- `make test` - Run tests with "unit" build tag
- `make test-coverage` - Run tests with coverage report
- `make test-report` - Generate test coverage report
- `make test-report-html` - Generate HTML test coverage report
- `make test-packages` - Dogfood test using the CLI (runs `./build/jx3-openapi-generation test`)

### Running a Single Test
```bash
go test -tags=unit -run TestFunctionName ./pkg/path/to/package/...
```

### Code Quality
- `make lint` - Run linting with golangci-lint
- `make fmt` - Format Go code and imports
- `make check` - Build and run tests (combines build + test)

### Utilities
- `make clean` - Clean build artifacts
- `make tidy-deps` - Clean up Go dependencies
- `make install` - Install binary to GOPATH/bin

## Architecture

### Core Components

**Command Structure:**
- `cmd/main.go` - Entry point that calls `cmd/app/main.go`
- `pkg/cmd/` - CLI command definitions using Cobra
- `pkg/cmd/generate/` - Main generate command and package generation logic
- `pkg/cmd/test/` - Test command for dogfooding with sensible defaults

**Package Generators:**
- `pkg/packageGenerator/base_generator.go` - Base generator with common functionality
- `pkg/packageGenerator/{language}/` - Language-specific generators (csharp, java, angular, etc.)
- Each generator extends BaseGenerator and implements language-specific packaging

**Generator Interface** (`pkg/domain/generators.go`):
```go
type PackageGenerator interface {
    GeneratePackage(outputDir string) (string, error)
    PushPackage(packageDir string) error
    GetPackageName() string
}
```

**Configuration:**
- `configs/{language}-openapitools.json` - OpenAPI generator configurations for each supported language
- `pkg/openapitools/config.go` - Configuration loading and management
- Language configs are loaded dynamically based on the requested output languages
- Config paths: `./configs` (local) → `/configs` (container fallback)

**Core Services:**
- `pkg/domain/` - Domain interfaces and types
- `pkg/file/file.go` - File I/O operations
- `pkg/git/git.go` - Git operations for pushing packages
- `pkg/scmClient/github/` - GitHub API integration
- `pkg/commandRunner/` - Command execution wrapper

### Workflow

1. CLI reads environment variables (VERSION, REPO_OWNER, REPO_NAME, SwaggerServiceName, SpecPath, etc.)
2. Validates OpenAPI specification exists at SpecPath
3. For each requested language, loads the appropriate config from `configs/`
4. Uses BaseGenerator to call `npx @openapitools/openapi-generator-cli` with language-specific config
5. Language-specific generators handle post-processing (copying templates, building packages)
6. Packages are pushed to configured repositories via Git/GitHub API (unless SKIP_PUSH=true)

## Environment Variables

Required for operation:
- `VERSION` - Package version
- `REPO_OWNER` - Repository owner
- `REPO_NAME` - Repository name
- `SwaggerServiceName` - Service name for package generation
- `SpecPath` - Path to OpenAPI spec file
- `GIT_USER` - Git username
- `GIT_TOKEN` - Git authentication token

Optional:
- `PackageName` - Package name (defaults to "Client")
- `SKIP_PUSH` - Set to "true" to skip pushing generated packages
- `OutputLanguages` - Space-separated list of languages (used in Jenkins X pipelines)

## Supported Languages

| Language   | Argument     | Notes                                    |
| ---------- | ------------ | ---------------------------------------- |
| C#         | `csharp`     |                                          |
| Java       | `java`       |                                          |
| Angular    | `angular`    |                                          |
| TypeScript | `typescript` |                                          |
| Python     | `python`     | **Not available for preview packages**   |
| Go         | `go`         | **Not available for preview packages**   |
| Rust       | `rust`       |                                          |

Go and Python use git-based repository management, so preview packages are not supported for these languages.

## Templates

Language-specific template files are stored in `templates/{language}/` for post-processing generated packages.

## Recent Upgrades

**OpenAPI Generator CLI**: Updated to v2.23.1 (from v2.13.4) across:
- Dockerfile: `@openapitools/openapi-generator-cli@2.23.1`
- CreateAngularPackage.sh: `@openapitools/openapi-generator-cli@2.23.1`
- Pipeline files: `yarn global add @openapitools/openapi-generator-cli@2.23.1`

**OpenAPI Generator Core**: Updated to v7.15.0 in all config files (`configs/*-openapitools.json`)

**Java Configuration Fixes**:
- Updated Java dependencies in `templates/java/build.gradle` for OpenAPI Generator 7.15.0 compatibility:
  - Added OAuth2 client: `org.apache.oltu.oauth2:org.apache.oltu.oauth2.client:1.0.2`
  - Added javax annotations: `javax.annotation:javax.annotation-api:1.3.2`
  - Updated all other dependencies to latest compatible versions
- Added Java generator properties: `library=okhttp-gson`, `serializationLibrary=gson`

## Testing

- `make test` - Run all unit tests with coverage (includes Java configuration validation)
- `make test-packages` - Dogfood test that generates packages using the CLI
- Java configuration tests in `pkg/packageGenerator/java/generator_test.go` validate the fixes for OpenAPI Generator 7.15.0
