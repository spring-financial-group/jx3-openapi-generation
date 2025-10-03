package javascript

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packagegenerator"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"path/filepath"
	"strings"
	"time"
)

const (
	packagingFilesDir = "/templates/javascript"
)

// Paths for use in generating angular packages
var (
	npmrcPath       = filepath.Join(packagingFilesDir, ".npmrc")
	packageJSONPath = filepath.Join(packagingFilesDir, "package.json")
)

// Packages installed by the generator
const (
	errNPMVersionAlreadyExists = "npm ERR! publish fail Cannot publish over existing version"
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
	packageDir, err := g.BaseGenerator.GeneratePackage(filepath.Join(outputDir, g.GetPackageName()), domain.Javascript)
	if err != nil {
		return "", err
	}

	err = g.Cmd.ExecuteAndLog(packageDir, "npm", "install")
	if err != nil {
		return "", errors.Wrap(err, "failed to run npm install")
	}

	err = g.Cmd.ExecuteAndLog(packageDir, "npm", "run", "build")
	if err != nil {
		return "", errors.Wrap(err, "failed to run npm build")
	}

	distDir := filepath.Join(packageDir, "dist")
	if err = g.FileIO.TemplateFiles(distDir, g, packageJSONPath, npmrcPath); err != nil {
		return "", err
	}
	return distDir, nil
}

func (g *Generator) GetPackageName() string {
	return fmt.Sprintf("%s-javascript", g.RepoName)
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
