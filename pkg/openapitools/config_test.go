package openapitools_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"spring-financial-group/jx3-openapi-generation/pkg/openapitools"
	"testing"
)

func TestConfig_GetConfig(t *testing.T) {
	config, err := openapitools.GetConfig("./../../openapitools.json")
	assert.NoError(t, err)
	fmt.Println(config)
}
