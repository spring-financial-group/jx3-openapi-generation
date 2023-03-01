package version_test

import (
	"github.com/stretchr/testify/assert"
	"spring-financial-group/jx3-openapi-generation/pkg/cmd/version"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := version.GetVersion()
	assert.NotEqual(t, "", version)
}
