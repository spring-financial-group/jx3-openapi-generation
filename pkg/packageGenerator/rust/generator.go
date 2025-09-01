package rust

import (
	"fmt"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"path/filepath"
)

const (
	packagingFilesDir = "/templates/rust"
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

	packageDir, err := g.BaseGenerator.GeneratePackage(filepath.Join(outputDir, g.GetPackageName()), domain.Rust)
	if err != nil {
		return "", err
	}

	if err = g.FileIO.TemplateFilesInDir(packagingFilesDir, packageDir, g); err != nil {
		return "", err
	}

	return packageDir, nil
}

func (g *Generator) setDynamicConfigVariables() {
	//g.Cfg.GeneratorCLI.Generators[domain.Rust].AdditionalProperties["packageName"] = g.GetPackageName()
}

func (g *Generator) GetPackageName() string {
	return fmt.Sprintf("mqube-%s-%s", g.ServiceName, g.PackageName)
}

func (g *Generator) PushPackage(packageDir string) error {
	//solutionPath := fmt.Sprintf("./src/%s/bin/Release/**/*.nupkg", g.GetPackageName())
	//return g.Cmd.ExecuteAndLog(packageDir, "dotnet", "nuget", "push", solutionPath, "-s", "mqube.packages", "--skip-duplicate")
	return nil
}
