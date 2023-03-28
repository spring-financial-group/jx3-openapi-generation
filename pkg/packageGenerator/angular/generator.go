package angular

import (
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
)

// Paths for use in generating angular packages
const (
	PackageJSONPath     = "./registry/package.json"
	TSConfigPath        = "./registry/tsconfig.json"
	ConfigurationTSPath = "/configuration.ts"
	NPMRCPath           = "./registry/.npmrc"
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

	_, err = g.getPackageJSON(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to get package.json")
	}

	err = g.FileIO.CopyManyToDir(packageDir, TSConfigPath, ConfigurationTSPath)
	if err != nil {
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

	// Copy the original package.json to the dist directory to remove the dependencies
	_, err = g.getPackageJSON(distDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to get package.json")
	}

	err = g.FileIO.CopyManyToDir(distDir, NPMRCPath, ConfigurationTSPath)
	if err != nil {
		return "", err
	}

	return distDir, nil
}

func (g *Generator) getPackageJSON(packageDir string) (string, error) {
	_, packageJSONPath, err := g.FileIO.CopyToDir(PackageJSONPath, packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to copy package.json")
	}
	err = g.FileIO.ReplaceInFile(packageJSONPath, "0.0.0", g.Version)
	if err != nil {
		return "", errors.Wrap(err, "failed to replace version in package.json")
	}
	return packageJSONPath, nil
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
	return g.ServiceName
}

func (g *Generator) PushPackage(packageDir string) error {
	return g.Cmd.ExecuteAndLog(packageDir, "npm", "publish")
}
