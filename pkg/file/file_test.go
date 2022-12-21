//go:build unit

package file_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/file"
	"spring-financial-group/jx3-openapi-generation/pkg/utils"
	"testing"
)

func TestFile_ReplaceInFile(t *testing.T) {
	fileIO := file.NewFileIO()

	testCases := []struct {
		name           string
		inputFile      string
		inputOld       string
		inputNew       string
		expectedOutput string
	}{
		{
			name: "Version",
			inputFile: `{
  "name": "@spring-financial-group/jx3-openapi-generation-angular",
  "version": "0.0.0",
}`,
			inputOld: "0.0.0",
			inputNew: "1.0.0",
			expectedOutput: `{
  "name": "@spring-financial-group/jx3-openapi-generation-angular",
  "version": "1.0.0",
}`,
		},
		{
			name: "RepoName",
			inputFile: `{
  "name": "@spring-financial-group/jx3-openapi-generation-angular",
  "version": "0.0.0",
}`,
			inputOld: "jx3-openapi-generation-angular",
			inputNew: "product-service",
			expectedOutput: `{
  "name": "@spring-financial-group/product-service",
  "version": "0.0.0",
}`,
		},
	}

	tmpDir := utils.CreateTestTmpDir(t, "file-replace-in-file")
	defer os.RemoveAll(tmpDir)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testPath := filepath.Join(tmpDir, tc.name+".json")
			err := os.WriteFile(testPath, []byte(tc.inputFile), 0755)
			assert.NoError(t, err)

			err = fileIO.ReplaceInFile(testPath, tc.inputOld, tc.inputNew)
			assert.NoError(t, err)

			actual, err := os.ReadFile(testPath)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, string(actual))
		})
	}
}
