package angular

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"path/filepath"
	"strings"
	"time"
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

	errNPMVersionAlreadyExists = "Cannot publish over existing version"
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
	out, err := g.Cmd.Execute(packageDir, "npm", "publish")
	log.Logger().Info(out)
	if err != nil {
		// NPM returns the error message on STDOUT, so we need to check there for the error
		if strings.Contains(out, errNPMVersionAlreadyExists) {
			log.Logger().Warnf("Package already exists at version %s, incrementing version and trying again", g.Version)
			err = g.incrementPackageVersion(packageDir)
			if err != nil {
				return err
			}
			return g.PushPackage(packageDir)
		}
		// Otherwise return the error
		return errors.Wrap(err, "failed to publish package")
	}
	return nil
}

func (g *Generator) incrementPackageVersion(packageDir string) error {
	currentV := g.Version
	newV := fmt.Sprintf("%s-%d", currentV, time.Now().Unix())

	log.Logger().Infof("Incrementing version %s to %s", currentV, newV)
	err := g.Cmd.ExecuteAndLog(packageDir, "npm", "version", newV)
	if err != nil {
		return errors.Wrapf(err, "failed to increment version to %s", newV)
	}
	g.Version = newV
	return nil
}
