package generate_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"spring-financial-group/jx3-openapi-generation/pkg/cmd/generate"
	"spring-financial-group/jx3-openapi-generation/pkg/domain/mocks"
	"testing"
)

func TestPackageOptions_Run(t *testing.T) {
	mockFactory := mocks.NewPackageGeneratorFactory(t)
	mockFileIO := mocks.NewFileIO(t)
	mockPackageGenerator := mocks.NewPackageGenerator(t)

	testCases := []struct {
		name           string
		inputLanguages []string
		pkgOpts        *generate.PackageOptions
	}{
		{
			name:           "OneLanguage",
			inputLanguages: []string{"java"},
			pkgOpts: &generate.PackageOptions{
				Options: &generate.Options{
					Version:            "1.0.0",
					SwaggerServiceName: "test-service",
					RepoOwner:          "spring-financial-group",
					RepoName:           "jx3-openapi-generation",
					SpecPath:           "/test/spec/path",
					FileIO:             mockFileIO,
				},
				GeneratorFactory: mockFactory,
			},
		},
		{
			name:           "MultipleLanguages",
			inputLanguages: []string{"java", "csharp"},
			pkgOpts: &generate.PackageOptions{
				Options: &generate.Options{
					Version:            "1.0.0",
					SwaggerServiceName: "test-service",
					RepoOwner:          "spring-financial-group",
					RepoName:           "jx3-openapi-generation",
					SpecPath:           "/test/spec/path",
					FileIO:             mockFileIO,
				},
				GeneratorFactory: mockFactory,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := "/tmp/package-generator"
			mockFileIO.On("MkTmpDir", "package-generator").Return(tmpDir, nil).Once()
			mockFileIO.On("CopyToWorkingDir", generate.OpenAPIToolsPath).Return(int64(0), nil).Once()

			for _, l := range tc.inputLanguages {
				mockFactory.On("NewGenerator", l).Return(mockPackageGenerator, nil).Once()

				outputDir := tmpDir + "/" + l
				mockFileIO.On("MkdirAll", outputDir, os.FileMode(0700)).Return(outputDir, nil).Once()

				mockPackageGenerator.On("GeneratePackage", tc.pkgOpts.SpecPath, outputDir).Return(outputDir, nil).Once()
				mockPackageGenerator.On("PushPackage", outputDir).Return(nil).Once()
			}

			mockFileIO.On("DeferRemove", tmpDir).Once()

			err := tc.pkgOpts.Run(tc.inputLanguages)
			assert.NoError(t, err)
		})
	}
}
