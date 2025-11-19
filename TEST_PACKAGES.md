# Package Generation Testing

This document describes how to test OpenAPI package generation locally and in CI.

## Overview

The package generation test validates that all supported language packages can be generated and built successfully from an OpenAPI specification. Build validation is performed during the generation process itself - each generator includes build steps:

- **Java**: `gradle publish`
- **C#**: `dotnet pack -c Release`  
- **Angular**: `npm install` + `ngc` compilation
- **TypeScript**: `npm install` + `npm run build`
- **Go**: `go mod tidy` + `go build` + mock generation
- **Python**: Package structure validation
- **Rust**: Cargo project setup validation

## Local Testing

### Option 1: Direct Script Execution
```bash
# Run locally with your current environment
make test-local
```

### Option 2: Docker Testing (Recommended)
```bash
# Build and run test in Docker container (replicates CI environment)
make test-docker

# Or run interactively for debugging
make test-docker-interactive
```

## Manual Script Execution

You can also run the test script directly:

```bash
./scripts/test-package-generation.sh
```

## Environment Variables

The test script uses these environment variables:

- `SwaggerServiceName` - Service name for package generation (default: OpenAPIPkgGenerationMock)
- `SpecPath` - Path to OpenAPI spec file (default: /swagger.json)
- `VERSION` - Package version (default: 0.0.0-TEST-SNAPSHOT)
- `REPO_OWNER` - Repository owner (default: spring-financial-group)
- `REPO_NAME` - Repository name (default: test-service)
- `PackageName` - Package name (default: Client)

## OpenAPI Specification

The test script will automatically create a minimal swagger.json file if one doesn't exist in the `mocks/` directory. For testing with your own specification, place it at `mocks/swagger.json`.

## Git Requirements

- **Go generator**: Requires git for repository operations (cloning, committing)
- **Other generators**: Work without git dependencies

If git is not available, the Go generator will be skipped automatically.

## CI Pipeline

The CI pipeline (`.lighthouse/jenkins-x/generate-test-packages.yaml`) uses the same test script to ensure consistency between local and CI environments.

## Troubleshooting

### Missing Dependencies
If you see errors about missing tools (gradle, dotnet, npm, etc.), the Docker test environment includes all required dependencies.

### Permission Errors
Ensure the test script is executable:
```bash
chmod +x ./scripts/test-package-generation.sh
```

### OpenAPI Specification Issues
The test includes a minimal valid OpenAPI 2.0 specification. If using your own spec:
- Ensure it's valid OpenAPI/Swagger format
- Place it at `mocks/swagger.json`
- The script will copy it to `/swagger.json` as expected by the tool