package domain

import "fmt"

const (
	CSharp     = "csharp"
	Java       = "java"
	Angular    = "angular"
	Python     = "python"
	Javascript = "javascript"
	Typescript = "typescript"
	Go         = "go"
	Rust       = "rust"
)

type PackageGenerator interface {
	// GeneratePackage generates a package from the given specification
	GeneratePackage(outputDir string) (string, error)
	// PushPackage pushes the generated package to the repository
	PushPackage(packageDir string) error
	// GetPackageName returns the name of the generated package
	GetPackageName() string
}

type UnsupportedLanguageError struct {
	Language string
}

func (e *UnsupportedLanguageError) Error() string {
	return fmt.Sprintf("unsupported language: %s", e.Language)
}
