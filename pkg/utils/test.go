package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func CreateTestTmpDir(t *testing.T, pattern string) string {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	tmpDir, err := os.MkdirTemp(wd, pattern)
	assert.NoError(t, err)
	return tmpDir
}
