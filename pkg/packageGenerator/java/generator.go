package java

import (
	"fmt"
	"github.com/pkg/errors"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"strings"
)

const (
	GradlePath = "./registry/build.gradle"
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

	packageDir, err := g.BaseGenerator.GeneratePackage(outputDir, domain.Java)
	if err != nil {
		return "", err
	}

	err = g.getBuildGradle(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to get build.gradle")
	}

	return packageDir, nil
}

func (g *Generator) setDynamicConfigVariables() {
	g.Cfg.GeneratorCLI.Generators[domain.Java].AdditionalProperties["basePackage"] = g.GetPackageName()
	g.Cfg.GeneratorCLI.Generators[domain.Java].AdditionalProperties["modelPackage"] = fmt.Sprintf("%s.models", g.GetPackageName())
}

func (g *Generator) getBuildGradle(packageDir string) error {
	_, buildGradlePath, err := g.FileIO.CopyToDir(GradlePath, packageDir)
	if err != nil {
		return errors.Wrap(err, "failed to copy build.gradle file")
	}
	err = g.FileIO.ReplaceInFile(buildGradlePath, "0.0.0", g.Version)
	if err != nil {
		return errors.Wrap(err, "failed to replace version in build.gradle")
	}
	return nil
}

func (g *Generator) GetPackageName() string {
	// Some PascalCase -> camelCase conversion
	pkgName := strings.ToLower(string(g.ServiceName[0]))
	pkgName += g.ServiceName[1:]
	return fmt.Sprintf("mqube.%s", pkgName)
}

func (g *Generator) PushPackage(packageDir string) error {
	return g.Cmd.ExecuteAndLog(packageDir, "gradle", "publish")
}
