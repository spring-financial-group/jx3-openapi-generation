package domain

import "fmt"

const (
	CSharp  = "csharp"
	Java    = "java"
	Angular = "angular"
	Python = "python"
)

type PackageGeneratorFactory interface {
	// NewGenerator returns a new PackageGenerator for the given language
	NewGenerator(language string) (PackageGenerator, error)
}

type PackageGenerator interface {
	// GeneratePackage generates a package from the given specification
	GeneratePackage(specificationPath, outputDir string) (string, error)
	// PushPackage pushes the generated package to the repository
	PushPackage(packageDir string) error
	// GetPackageName returns the name of the generated package
	GetPackageName() string
}

type ErrUnsupportedLanguage struct {
	Language string
}

func (e *ErrUnsupportedLanguage) Error() string {
	return fmt.Sprintf("unsupported language: %s", e.Language)
}
