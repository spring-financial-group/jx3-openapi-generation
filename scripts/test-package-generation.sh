#!/bin/bash
set -e

# Script for testing OpenAPI package generation with build validation
# Used both in CI pipeline and local Docker testing

echo "🔧 OpenAPI Package Generation Test Script"
echo "=========================================="

# Check if running in pipeline or local environment
if [ -n "$CI" ] || [ -n "$TEKTON_RUN_ID" ]; then
    echo "📍 Running in CI/Pipeline environment"
    WORK_DIR="/workspace/source"
else
    echo "📍 Running in local environment"
    WORK_DIR=$(pwd)
fi

echo "Working directory: $WORK_DIR"

# Setup git if available (required for Go generator, optional for others)
if command -v git >/dev/null 2>&1; then
    echo "🔧 Configuring git..."
    git config --global user.name "${GIT_AUTHOR_NAME:-mqube-bot}" 2>/dev/null || true
    git config --global user.email "${GIT_AUTHOR_EMAIL:-mqube-bot@mqube.com}" 2>/dev/null || true
    LANGUAGES="go python angular csharp java typescript rust"
else
    echo "⚠️  Git not available - skipping Go generator (requires git operations)"
    LANGUAGES="python angular csharp java typescript rust"
fi

# Build the binary if not exists
echo "🔨 Building jx3-openapi-generation binary..."
cd "$WORK_DIR"

if [ ! -f "./build/jx3-openapi-generation" ]; then
    echo "Binary not found, building..."
    make build
fi

# Copy binary to mocks directory for testing
echo "📋 Preparing test environment..."
mkdir -p ./mocks
cp ./build/jx3-openapi-generation ./mocks/

# Ensure swagger.json exists for testing
if [ ! -f "./mocks/swagger.json" ]; then
    echo "⚠️  swagger.json not found in mocks directory"
    echo "Creating a minimal swagger.json for testing..."
    cat > ./mocks/swagger.json << 'EOF'
{
  "swagger": "2.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "host": "api.example.com",
  "basePath": "/v1",
  "schemes": ["https"],
  "paths": {
    "/test": {
      "get": {
        "summary": "Test endpoint",
        "responses": {
          "200": {
            "description": "Success"
          }
        }
      }
    }
  }
}
EOF
fi

# Set environment variables required by the tool
export SwaggerServiceName=${SwaggerServiceName:-OpenAPIPkgGenerationMock}
export SpecPath=${SpecPath:-/swagger.json}
export VERSION=${VERSION:-0.0.0-TEST-SNAPSHOT}
export REPO_OWNER=${REPO_OWNER:-spring-financial-group}
export REPO_NAME=${REPO_NAME:-test-service}
export PackageName=${PackageName:-Client}
export GIT_USER=${GIT_USER:-test-user}
export GIT_TOKEN=${GIT_TOKEN:-test-token}
export SKIP_PUSH=${SKIP_PUSH:-true}

# Copy swagger.json and configs to expected locations (handle read-only filesystem)
if [ -w / ]; then
    cp ./mocks/swagger.json /swagger.json
    # Copy configs directory to expected location if not exists
    if [ ! -d /configs ]; then
        cp -r ./configs /configs
    fi
else
    # If root filesystem is read-only, use relative paths
    echo "Root filesystem is read-only, using alternative locations"
    export SpecPath="./mocks/swagger.json"
    # Create configs symlink if possible, otherwise skip languages that need it
    if [ ! -L /configs ] && [ ! -d /configs ]; then
        ln -sf "$(pwd)/configs" /configs 2>/dev/null || echo "⚠️  Cannot create /configs symlink - some generators may fail"
    fi
fi

echo "🔄 Environment variables:"
echo "  SwaggerServiceName: $SwaggerServiceName"
echo "  SpecPath: $SpecPath"
echo "  VERSION: $VERSION"
echo "  REPO_OWNER: $REPO_OWNER"
echo "  REPO_NAME: $REPO_NAME"

# Generate packages with build validation
echo ""
echo "🚀 Generating packages with build validation..."
echo "Languages: $LANGUAGES"
echo ""

# Verify setup
echo "📍 Running from: $(pwd)"
echo "📂 SpecPath: $SpecPath"

# Run the package generation from the project root (where configs are located)
# Each generator includes its own build validation:
# - Java: gradle publish
# - C#: dotnet pack  
# - Angular: npm install + ngc
# - TypeScript: npm install + npm run build
# - Go: go mod tidy + go build
# - Python: package structure validation
# - Rust: cargo setup validation

set -x  # Enable tracing
echo "🔧 Invoking package generation with languages: $LANGUAGES"
./mocks/jx3-openapi-generation generate packages $LANGUAGES
exit_code=$?
set +x  # Disable tracing

if [ $exit_code -ne 0 ]; then
    echo ""
    echo "❌ Package generation failed with exit code $exit_code"
    exit $exit_code
else
    echo ""
    echo "✅ All packages generated successfully!"
    echo "🎉 Package generation test completed!"
fi