#!/bin/bash

# Finds the spec file from the current directory and all subdirectories
# If none is found, return the default path
DEFAULT_PATH=$1

SWAG_PATH=$(find . -type f -name "swagger.json")
OPENAPI_PATH=$(find . -type f -name "openapi.json")
if test "$SWAG_PATH" ; then
  echo "Swagger specification found at $SWAG_PATH"
  SPEC_PATH=$SWAG_PATH
elif test "$OPENAPI_PATH" ; then
  echo "OpenAPI specification found at $OPENAPI_PATH"
  SPEC_PATH=$OPENAPI_PATH
else
  echo "Specification not found in repository using $DEFAULT_PATH"
  SPEC_PATH=$DEFAULT_PATH
fi

echo "$SPEC_PATH"
