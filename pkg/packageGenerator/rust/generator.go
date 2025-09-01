package rust

import (
	"path/filepath"

	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
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
	g.Cfg.GeneratorCLI.Generators[domain.Rust].AdditionalProperties["packageName"] = g.GetPackageName()
	g.Cfg.GeneratorCLI.Generators[domain.Rust].AdditionalProperties["packageVersion"] = g.Version
}

func (g *Generator) GetPackageName() string {
	return g.RepoName
}

func (g *Generator) PushPackage(packageDir string) error {
	//solutionPath := fmt.Sprintf("./src/%s/bin/Release/**/*.nupkg", g.GetPackageName())
	//return g.Cmd.ExecuteAndLog(packageDir, "dotnet", "nuget", "push", solutionPath, "-s", "mqube.packages", "--skip-duplicate")
	return nil
}
