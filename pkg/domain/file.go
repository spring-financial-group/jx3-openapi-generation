package domain

import (
	"fmt"
	"os"
	"strings"
)

type FileIO interface {
	// Read reads the file at the given path and returns its content as a byte array
	Read(path string) ([]byte, error)
	// Write writes the given data to the file at the given path
	Write(path string, data []byte, perm os.FileMode) error
	// Exists checks if a file exists
	Exists(path string) (bool, error)
	// Copy copies a file from the source to the target
	Copy(src string, dst string) (int64, error)
	// CopyToDir copies a file from the source to the target directory
	// preserving the file name
	CopyToDir(srcPath string, dstDir string) (int64, string, error)
	// CopyToWorkingDir copies a file from the source to the current working directory
	// preserving the file name
	CopyToWorkingDir(srcPath string) (int64, error)
	// CopyManyToDir copies multiple files from the source to the target directory
	// preserving the file names
	CopyManyToDir(dstDir string, srcFiles ...string) error
	// Move moves a file from the source to the target
	Move(src string, dst string) error
	// MkdirAll creates a directory and all its parents with the given permissions
	MkdirAll(path string, perm os.FileMode) (string, error)
	// MkTmpDir creates a temporary directory with the given prefix and returns its path
	MkTmpDir(prefix string) (string, error)
	// Remove removes a file or directory
	Remove(path string) error
	// DeferRemove removes a file or directory logging an error if it fails
	DeferRemove(path string)
	// ReplaceInFile replaces the given string in the file at the given path
	ReplaceInFile(path, old, new string) error
	// TemplateFiles renders the given files using the given object and writes them to the given directory
	TemplateFiles(dstDir string, obj any, filePaths ...string) error
	// TemplateFilesInDir renders any files in the given source directory using the given object and writes them to the
	// given destination directory using the same file names
	TemplateFilesInDir(srcDir, dstDir string, obj any) error
}

type ErrFileNotFound struct {
	FilePath string
}

func (f *ErrFileNotFound) Error() string {
	return fmt.Sprintf("file not found: %s", f.FilePath)
}

type ErrEnvironmentVariableNotFound struct {
	VariableNames []string
}

func (e *ErrEnvironmentVariableNotFound) Error() string {
	variableNamesCommaSeparated := strings.Join(e.VariableNames, ", ")
	return fmt.Sprintf("environment variables not found: %s", variableNamesCommaSeparated)
}
