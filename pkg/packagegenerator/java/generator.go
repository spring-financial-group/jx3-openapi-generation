package java

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packagegenerator"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
)

const (
	packagingFilesDir = "/templates/java"
)

type Generator struct {
	*packagegenerator.BaseGenerator
}

func NewGenerator(baseGenerator *packagegenerator.BaseGenerator) *Generator {
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
	g.Cfg.GeneratorCLI.Generators[domain.Java].AdditionalProperties["basePackage"] = g.getModelName()
	g.Cfg.GeneratorCLI.Generators[domain.Java].AdditionalProperties["modelPackage"] = fmt.Sprintf("%s.models", g.getModelName())
}

func (g *Generator) GetPackageName() string {
	// Replace first hyphen with a dot mqube-foo-service -> mqube.foo-service
	return strings.Replace(g.RepoName, "-", ".", 1)
}

func (g *Generator) getModelName() string {
	// convert pascal case to camel case
	return fmt.Sprintf("mqube.%s", utils.FirstCharToLower(g.ServiceName))
}

func (g *Generator) PushPackage(packageDir string) error {
	return g.Cmd.ExecuteAndLog(packageDir, "gradle", "publish")
}
