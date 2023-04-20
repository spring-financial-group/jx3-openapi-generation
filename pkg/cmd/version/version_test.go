package version_test

import (
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/cmd/version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := version.GetVersion()
	assert.NotEqual(t, "", version)
}
