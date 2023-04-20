package java

import (
	"fmt"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"strings"
)

const (
	packagingFilesDir = "/templates/java"
)

type Generator struct {
	*packageGenerator.BaseGenerator
}

func NewGenerator(baseGenerator *packageGenerator.BaseGenerator) *Generator {
	return &Generator{
		BaseGenerator: baseGenerator,
	}
}

func (g *Generator) GeneratePackage(outputDir string) (string, error) {
	g.setDynamicConfigVariables()

	packageDir, err := g.BaseGenerator.GeneratePackage(filepath.Join(outputDir, g.GetPackageName()), domain.Java)
	if err != nil {
		return "", err
	}

	if err = g.FileIO.TemplateFilesInDir(packagingFilesDir, packageDir, g); err != nil {
		return "", err
	}

	return packageDir, nil
}

func (g *Generator) setDynamicConfigVariables() {
	g.Cfg.GeneratorCLI.Generators[domain.Java].AdditionalProperties["basePackage"] = g.GetPackageName()
	g.Cfg.GeneratorCLI.Generators[domain.Java].AdditionalProperties["modelPackage"] = fmt.Sprintf("%s.models", g.GetPackageName())
}

func (g *Generator) GetPackageName() string {
	// Replace first hyphen with a dot mqube-foo-service -> mqube.foo-service
	return strings.Replace(g.RepoName, "-", ".", 1)
}

func (g *Generator) PushPackage(packageDir string) error {
	return g.Cmd.ExecuteAndLog(packageDir, "gradle", "publish")
}
