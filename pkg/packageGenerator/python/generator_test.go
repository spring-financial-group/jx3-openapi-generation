package python_test

import (
	"fmt"
	"github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"spring-financial-group/jx3-openapi-generation/pkg/domain/mocks"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/python"
	"spring-financial-group/jx3-openapi-generation/pkg/utils"
	"testing"
)

func TestGenerator_GeneratePackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)
	mockFileIO := mocks.NewFileIO(t)
	mockGit := mocks.NewGitter(t)

	previousPackages := map[string]python.PackageInfo{
		"test-service": {
			Directory: "/tmp/python-generator/test-service",
			Name: "test-service",
			Version: "1.0.0",
		},
		"other-test-service": {
			Directory: "/tmp/python-generator/other-test-service",
			Name: "other-test-service",
			Version: "1.0.0",
		},
	}

	data, err := utils.MarshalJSON(previousPackages)
	assert.NoError(t, err)


	testCases := []struct {
		name   string
		pkgGen *python.Generator
	}{
		{
			name: "1",
			pkgGen: &python.Generator{
				Version:     "1.0.0",
				ServiceName: "test-service",
				RepoOwner:   "test-owner",
				RepoName:    "test-repo",
				Cmd:         mockCmdRunner,
				FileIO:      mockFileIO,
				Git:         mockGit,
			},
		},
		{
			name: "2",
			pkgGen: &python.Generator{
				Version:     "12.5.0",
				ServiceName: "other-test-service",
				RepoOwner:   "other-test-owner",
				RepoName:    "other-test-repo",
				Cmd:         mockCmdRunner,
				FileIO:      mockFileIO,
				Git:         mockGit,
			},
		},
	}

	outputDir := "/tmp/python-generator/python-service"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generatorDir := fmt.Sprintf("%s/%s", outputDir, tc.pkgGen.GetPackageName())
			mockFileIO.On("MkdirAll", generatorDir, os.FileMode(0755)).Return(generatorDir, nil).Once()

			repoDir := fmt.Sprintf("%s/%s", outputDir, tc.pkgGen.RepoName)
			mockGit.On("Clone", generatorDir, python.PipelineSchemasURL).Return(repoDir, nil).Once()
			mockGit.On("CheckoutBranch", repoDir, fmt.Sprintf("update/%s/%s", tc.pkgGen.GetPackageName(), tc.pkgGen.Version)).Return(nil).Once()

			packageDir := fmt.Sprintf("%s/%s", repoDir, tc.pkgGen.GetPackageName())
			mockFileIO.On("MkdirAll", packageDir, os.FileMode(0755)).Return(packageDir, nil).Once()

			mockCmdRunner.On("ExecuteAndLog", packageDir, "datamodel-codegen", "--input", "spec.json",
				"--input-file-type", "auto", "--output", "schemas.py",
			).Return(nil).Once()

			initPy := fmt.Sprintf("%s/__init__.py", packageDir)
			mockFileIO.On("Write", initPy, []byte{}, os.FileMode(0755)).Return(nil).Once()

			packageJSONPath := fmt.Sprintf("%s/packages.json", repoDir)
			mockFileIO.On("Read", packageJSONPath).Return(data, nil).Once()

			mockFileIO.On("Write", packageJSONPath, mock.AnythingOfType("[]uint8"), os.FileMode(0755)).Return(nil).Once()

			schemasPy := fmt.Sprintf("%s/schemas.py", packageDir)
			mockGit.On("AddFiles", repoDir, schemasPy, initPy, packageJSONPath).Return(nil).Once()

			mockGit.On("Commit", repoDir, fmt.Sprintf("chore(deps): upgrade %s package -> %s", tc.pkgGen.GetPackageName(), tc.pkgGen.Version)).Return(nil).Once()

			_, err := tc.pkgGen.GeneratePackage("spec.json", outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_PushPackage(t *testing.T) {
	mockCmdRunner := mocks.NewCommandRunner(t)
	mockGit := mocks.NewGitter(t)
	mockScm := mocks.NewScmClient(t)

	testCases := []struct {
		name   string
		pkgGen *python.Generator
	}{
		{
			name: "1",
			pkgGen: &python.Generator{
				Version:     "1.0.0",
				ServiceName: "test-service",
				RepoOwner:   "test-owner",
				RepoName:    "test-repo",
				Cmd:         mockCmdRunner,
				Git: mockGit,
				Scm: mockScm,
			},
		},
		{
			name: "2",
			pkgGen: &python.Generator{
				Version:     "12.5.0",
				ServiceName: "other-test-service",
				RepoOwner:   "other-test-owner",
				RepoName:    "other-test-repo",
				Cmd:         mockCmdRunner,
				Git:         mockGit,
				Scm:         mockScm,
			},
		},
	}

	outputDir := "/tmp/python-generator/python-service"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentBranch := "update"
			mockGit.On("GetCurrentBranch", outputDir).Return(currentBranch, nil).Once()
			mockGit.On("Push", outputDir, currentBranch).Return( nil).Once()
			defaultBranch := "origin/master"
			mockGit.On("GetDefaultBranchName", outputDir).Return(defaultBranch, nil).Once()

			newPR := &github.NewPullRequest{
				Title:               utils.NewPtr(fmt.Sprintf("chore(deps): upgrade %s package -> %s", tc.pkgGen.GetPackageName(), tc.pkgGen.Version)),
				Head:                &currentBranch,
				Base:                utils.NewPtr("master"),
				Body:                utils.NewPtr(fmt.Sprintf("Automated python schemas update for %s", tc.pkgGen.GetPackageName())),
				MaintainerCanModify: utils.NewPtr(true),
			}
			returnedPR := &github.PullRequest{
				Number: utils.NewPtr(1),
			}
			mockScm.On("CreatePullRequest", mock.AnythingOfType("*context.emptyCtx"), newPR).Return(returnedPR, nil).Once()
			mockScm.On("RequestReviewers", mock.AnythingOfType("*context.emptyCtx"), []string{"Reton2", "stelios93"}, 1).Return(nil, nil).Once()
			mockScm.On("AddLabels", mock.AnythingOfType("*context.emptyCtx"), []string{"updatebot"}, 1).Return(nil, nil).Once()

			err := tc.pkgGen.PushPackage(outputDir)
			assert.NoError(t, err)
		})
	}
}

func TestGenerator_GetPackageName(t *testing.T) {
	testCases := []struct {
		name                string
		pkgGen              *python.Generator
		expectedPackageName string
	}{
		{
			name: "1",
			pkgGen: &python.Generator{
				RepoName: "test-service",
			},
			expectedPackageName: "test_service",
		},
		{
			name: "2",
			pkgGen: &python.Generator{
				RepoName: "other-test-service",
			},
			expectedPackageName: "other_test_service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPackageName := tc.pkgGen.GetPackageName()
			assert.Equal(t, tc.expectedPackageName, actualPackageName)
		})
	}
}
