package java_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"spring-financial-group/jx3-openapi-generation/pkg/domain/mocks"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/java"
	"testing"
)

func TestGenerator_GeneratePackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)
	mockFileIO := mocks.NewFileIO(t)

	testCases := []struct {
		name   string
		pkgGen *java.Generator
	}{
		{
			name: "1",
			pkgGen: &java.Generator{
				Version:     "1.0.0",
				ServiceName: "test-service",
				RepoOwner:   "test-owner",
				RepoName:    "test-repo",
				Cmd:         mockCmdRunner,
				FileIO:      mockFileIO,
			},
		},
		{
			name: "2",
			pkgGen: &java.Generator{
				Version:     "12.5.0",
				ServiceName: "other-test-service",
				RepoOwner:   "other-test-owner",
				RepoName:    "other-test-repo",
				Cmd:         mockCmdRunner,
				FileIO:      mockFileIO,
			},
		},
	}

	outputDir := "/tmp/java-generator/java-service"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packageDir := fmt.Sprintf("%s/%s", outputDir, tc.pkgGen.GetPackageName())
			mockFileIO.On("MkdirAll", packageDir, os.FileMode(0755)).Return(packageDir, nil).Once()
			mockCmdRunner.On("ExecuteAndLog", outputDir, "npx", "openapi-generator-cli", "generate", "-i", "spec.json", "-g", "java", "-o", packageDir, "--git-user-id", tc.pkgGen.RepoOwner, "--git-repo-id", tc.pkgGen.RepoName, "--global-property", "models,modelTests=false,modelDocs=false", fmt.Sprintf("-p basePackage=%s -p modelPackage=%s.models", tc.pkgGen.GetPackageName(), tc.pkgGen.GetPackageName()), "-p", "dateLibrary=java8-localdatetime").Return(nil).Once()
			mockFileIO.On("CopyToDir", java.GradlePath, packageDir).Return(int64(0), packageDir+"/build.gradle", nil).Once()
			mockFileIO.On("ReplaceInFile", packageDir+"/build.gradle", "0.0.0", tc.pkgGen.Version).Return(nil).Once()

			_, err := tc.pkgGen.GeneratePackage("spec.json", outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_PushPackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)

	testCases := []struct {
		name   string
		pkgGen *java.Generator
	}{
		{
			name: "1",
			pkgGen: &java.Generator{
				Version:     "1.0.0",
				ServiceName: "test-service",
				RepoOwner:   "test-owner",
				RepoName:    "test-repo",
				Cmd:         mockCmdRunner,
			},
		},
		{
			name: "2",
			pkgGen: &java.Generator{
				Version:     "12.5.0",
				ServiceName: "other-test-service",
				RepoOwner:   "other-test-owner",
				RepoName:    "other-test-repo",
				Cmd:         mockCmdRunner,
			},
		},
	}

	outputDir := "/tmp/java-generator/java-service"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCmdRunner.On("ExecuteAndLog", outputDir, "gradle", "publish").Once().Return(nil)

			err := tc.pkgGen.PushPackage(outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_GetPackageName(t *testing.T) {
	testCases := []struct {
		name                string
		pkgGen              *java.Generator
		expectedPackageName string
	}{
		{
			name: "1",
			pkgGen: &java.Generator{
				ServiceName: "CaseService",
			},
			expectedPackageName: "mqube.caseService",
		},
		{
			name: "2",
			pkgGen: &java.Generator{
				ServiceName: "ProductService",
			},
			expectedPackageName: "mqube.productService",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPackageName := tc.pkgGen.GetPackageName()
			assert.Equal(t, tc.expectedPackageName, actualPackageName)
		})
	}
}
