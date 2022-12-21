package packageGenerator_test

import (
	"github.com/stretchr/testify/assert"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/angular"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/csharp"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/java"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/python"
	"testing"
)

func TestFactory_NewGenerator(t *testing.T) {
	testFactory := *packageGenerator.NewFactory("1.0.0", "test-service", "spring-financial-group", "jx3-openapi-generation", "test_token")

	testCases := []struct {
		name          string
		inputLang     string
		expectedType  any
		isErrExpected bool
	}{
		{
			name:          "CSharp",
			inputLang:     "csharp",
			expectedType:  &csharp.Generator{},
			isErrExpected: false,
		},
		{
			name:          "Java",
			inputLang:     "java",
			expectedType:  &java.Generator{},
			isErrExpected: false,
		},
		{
			name:          "Angular",
			inputLang:     "angular",
			expectedType:  &angular.Generator{},
			isErrExpected: false,
		},
		{
			name:          "Python",
			inputLang:     "python",
			expectedType:  &python.Generator{},
			isErrExpected: false,
		},
		{
			name:          "UnsupportedLanguage",
			inputLang:     "fortran",
			expectedType:  nil,
			isErrExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualGenerator, err := testFactory.NewGenerator(tc.inputLang)
			if !tc.isErrExpected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, &domain.ErrUnsupportedLanguage{}, err)
			}
			assert.IsType(t, tc.expectedType, actualGenerator)
		})
	}
}
