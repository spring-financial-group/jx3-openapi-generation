package csharp

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
)

const (
	packagingFilesDir = "/templates/csharp/nuget.config"
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

	packageDir, err := g.BaseGenerator.GeneratePackage(filepath.Join(outputDir, g.GetPackageName()), domain.CSharp)
	if err != nil {
		return "", err
	}

	if err = g.FileIO.TemplateFilesInDir(packagingFilesDir, packageDir, g); err != nil {
		return "", err
	}

	err = g.Cmd.ExecuteAndLog(packageDir, "dotnet", "pack", "-c", "Release", fmt.Sprintf("-p:VERSION=%s", g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to pack solution")
	}
	return packageDir, nil
}

func (g *Generator) setDynamicConfigVariables() {
	g.Cfg.GeneratorCLI.Generators[domain.CSharp].AdditionalProperties["packageName"] = g.GetPackageName()
}

func (g *Generator) GetPackageName() string {
	return fmt.Sprintf("Mqube.%s.Client", g.ServiceName)
}

func (g *Generator) PushPackage(packageDir string) error {
	solutionPath := fmt.Sprintf("./src/%s/bin/Release/**/*.nupkg", g.GetPackageName())
	return g.Cmd.ExecuteAndLog(packageDir, "dotnet", "nuget", "push", solutionPath, "-s", "mqube.packages", "--skip-duplicate")
}
