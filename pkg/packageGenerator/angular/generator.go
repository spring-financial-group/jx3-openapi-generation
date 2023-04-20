package angular

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
)

const (
	packagingFilesDir = "/templates/angular"
)

// Paths for use in generating angular packages
var (
	npmrcPath       = filepath.Join(packagingFilesDir, ".npmrc")
	packageJSONPath = filepath.Join(packagingFilesDir, "package.json")
	tsConfigPath    = filepath.Join(packagingFilesDir, "tsconfig.json")
)

// Packages installed by the generator
const (
	RXJS          = "rxjs@6.6.7"
	Zone          = "zone.js@0.9.1"
	AngularCore   = "@angular/core@8.2.14"
	AngularCommon = "@angular/common@8.2.14"
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
	packageDir, err := g.BaseGenerator.GeneratePackage(filepath.Join(outputDir, g.GetPackageName()), domain.Angular)
	if err != nil {
		return "", err
	}

	if err = g.FileIO.TemplateFiles(packageDir, g, packageJSONPath, tsConfigPath); err != nil {
		return "", err
	}

	err = g.installNPMPackages(packageDir, RXJS, Zone, AngularCore, AngularCommon)
	if err != nil {
		return "", err
	}

	err = g.Cmd.ExecuteAndLog(packageDir, "ngc")
	if err != nil {
		return "", errors.Wrap(err, "failed to run ngc")
	}

	distDir := filepath.Join(outputDir, "dist")
	if err = g.FileIO.TemplateFiles(distDir, g, packageJSONPath, npmrcPath); err != nil {
		return "", err
	}
	return distDir, nil
}

func (g *Generator) installNPMPackages(dir string, packages ...string) error {
	for _, pkg := range packages {
		err := g.Cmd.ExecuteAndLog(dir, "npm", "install", "--save", pkg)
		if err != nil {
			return errors.Wrapf(err, "failed to install %s", pkg)
		}
	}
	return nil
}

func (g *Generator) GetPackageName() string {
	return fmt.Sprintf("%s-angular", g.RepoName)
}

func (g *Generator) PushPackage(packageDir string) error {
	return g.Cmd.ExecuteAndLog(packageDir, "npm", "publish")
}
