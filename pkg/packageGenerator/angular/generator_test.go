package angular_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/domain/mocks"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/angular"
	"testing"
)

func TestGenerator_GeneratePackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)
	mockFileIO := mocks.NewFileIO(t)

	testCases := []struct {
		name   string
		pkgGen *angular.Generator
	}{
		{
			name: "1",
			pkgGen: &angular.Generator{
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
			pkgGen: &angular.Generator{
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

			mockCmdRunner.On("ExecuteAndLog", outputDir, "npx", "openapi-generator-cli", "generate",
				"-i", "spec.json", "-g", "typescript-angular", "-o", packageDir,
				"--additional-properties=fileNaming=camelCase,ngVersion=10.0.0,stringEnums=true",
				"--enable-post-process-file", "--remove-operation-id-prefix").Return(nil).Once()

			pkgJSONPath := fmt.Sprintf("%s/package.json", packageDir)
			mockFileIO.On("CopyToDir", angular.PackageJSONPath, packageDir).Return(int64(0), pkgJSONPath, nil).Once()
			mockFileIO.On("ReplaceInFile", pkgJSONPath, "0.0.0", tc.pkgGen.Version).Return(nil).Once()

			mockFileIO.On("CopyManyToDir", packageDir, angular.TSConfigPath, angular.ConfigurationTSPath).Return(nil).Once()

			for _, pkg := range []string{angular.RXJS, angular.Zone, angular.AngularCore, angular.AngularCommon} {
				mockCmdRunner.On("ExecuteAndLog", packageDir, "npm", "install", "--save", pkg).Return(nil).Once()
			}

			mockCmdRunner.On("ExecuteAndLog", packageDir, "ngc").Return(nil).Once()

			distDir := filepath.Join(outputDir, "dist")
			mockFileIO.On("CopyManyToDir", distDir, angular.NPMRCPath, angular.ConfigurationTSPath, pkgJSONPath).Return(nil).Once()

			_, err := tc.pkgGen.GeneratePackage("spec.json", outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_PushPackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)

	testCases := []struct {
		name   string
		pkgGen *angular.Generator
	}{
		{
			name: "1",
			pkgGen: &angular.Generator{
				Version:     "1.0.0",
				ServiceName: "test-service",
				RepoOwner:   "test-owner",
				RepoName:    "test-repo",
				Cmd:         mockCmdRunner,
			},
		},
		{
			name: "2",
			pkgGen: &angular.Generator{
				Version:     "12.5.0",
				ServiceName: "other-test-service",
				RepoOwner:   "other-test-owner",
				RepoName:    "other-test-repo",
				Cmd:         mockCmdRunner,
			},
		},
	}

	outputDir := "/tmp/angular-generator/angular-service"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCmdRunner.On("ExecuteAndLog", outputDir, "npm", "publish").Once().Return(nil)

			err := tc.pkgGen.PushPackage(outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_GetPackageName(t *testing.T) {
	testCases := []struct {
		name                string
		pkgGen              *angular.Generator
		expectedPackageName string
	}{
		{
			name: "1",
			pkgGen: &angular.Generator{
				ServiceName: "test-service",
			},
			expectedPackageName: "test-service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPackageName := tc.pkgGen.GetPackageName()
			assert.Equal(t, tc.expectedPackageName, actualPackageName)
		})
	}
}
