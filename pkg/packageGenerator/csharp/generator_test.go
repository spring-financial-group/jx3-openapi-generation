package csharp_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"spring-financial-group/jx3-openapi-generation/pkg/domain/mocks"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/csharp"
	"testing"
)

func TestGenerator_GeneratePackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)
	mockFileIO := mocks.NewFileIO(t)

	testCases := []struct {
		name   string
		pkgGen *csharp.Generator
	}{
		{
			name: "1",
			pkgGen: &csharp.Generator{
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
			pkgGen: &csharp.Generator{
				Version:     "12.5.0",
				ServiceName: "other-test-service",
				RepoOwner:   "other-test-owner",
				RepoName:    "other-test-repo",
				Cmd:         mockCmdRunner,
				FileIO:      mockFileIO,
			},
		},
	}

	outputDir := "/tmp/csharp-generator/csharp-service"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packageDir := fmt.Sprintf("%s/%s", outputDir, tc.pkgGen.GetPackageName())
			mockFileIO.On("MkdirAll", packageDir, os.FileMode(0755)).Return(packageDir, nil).Once()

			mockCmdRunner.On("ExecuteAndLog", outputDir, "npx", "openapi-generator-cli", "generate", "-i", "spec.json", "-g", "csharp-netcore", "-o", packageDir, "--git-user-id", tc.pkgGen.RepoOwner, "--git-repo-id", tc.pkgGen.RepoName, fmt.Sprintf("--additional-properties=targetFramework=netstandard2.1,packageName=%s,packageVersion=%s,netCoreProjectFile=true,optionalEmitDefaultValues=true,validatable=false", tc.pkgGen.GetPackageName(), tc.pkgGen.Version)).Return(nil).Once()

			mockFileIO.On("CopyToDir", csharp.NugetConfigPath, packageDir).Return(int64(0), "", nil).Once()

			mockCmdRunner.On("ExecuteAndLog", packageDir, "dotnet", "pack", "-c", "Release", fmt.Sprintf("-p:VERSION=%s", tc.pkgGen.Version)).Return(nil).Once()

			_, err := tc.pkgGen.GeneratePackage("spec.json", outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_PushPackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)

	testCases := []struct {
		name   string
		pkgGen *csharp.Generator
	}{
		{
			name: "1",
			pkgGen: &csharp.Generator{
				Version:     "1.0.0",
				ServiceName: "test-service",
				RepoOwner:   "test-owner",
				RepoName:    "test-repo",
				Cmd:         mockCmdRunner,
			},
		},
		{
			name: "2",
			pkgGen: &csharp.Generator{
				Version:     "12.5.0",
				ServiceName: "other-test-service",
				RepoOwner:   "other-test-owner",
				RepoName:    "other-test-repo",
				Cmd:         mockCmdRunner,
			},
		},
	}

	outputDir := "/tmp/csharp-generator/csharp-service"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCmdRunner.On("ExecuteAndLog", outputDir, "dotnet", "nuget", "push", fmt.Sprintf("./src/%s/bin/Release/**/*.nupkg", tc.pkgGen.GetPackageName()), "-s", "mqube.packages", "--skip-duplicate").Once().Return(nil)

			err := tc.pkgGen.PushPackage(outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_GetPackageName(t *testing.T) {
	testCases := []struct {
		name                string
		pkgGen              *csharp.Generator
		expectedPackageName string
	}{
		{
			name: "1",
			pkgGen: &csharp.Generator{
				ServiceName: "test-service",
			},
			expectedPackageName: "Mqube.test-service.Client",
		},
		{
			name: "2",
			pkgGen: &csharp.Generator{
				ServiceName: "other-test-service",
			},
			expectedPackageName: "Mqube.other-test-service.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPackageName := tc.pkgGen.GetPackageName()
			assert.Equal(t, tc.expectedPackageName, actualPackageName)
		})
	}
}
