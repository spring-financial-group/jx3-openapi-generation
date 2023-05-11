package utils_test

import (
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFirstCharToLower(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "A",
			expected: "a",
		},
		{
			name:     "multiple characters",
			input:    "ABC",
			expected: "aBC",
		},
		{
			name:     "service name",
			input:    "FooService",
			expected: "fooService",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := utils.FirstCharToLower(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
