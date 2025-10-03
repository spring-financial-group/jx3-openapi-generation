package version_test

import (
	"testing"

	"github.com/spring-financial-group/jx3-openapi-generation/pkg/cmd/version"
	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	assert.NotEmpty(t, version.GetVersion())
}
