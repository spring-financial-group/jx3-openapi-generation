package angular

import (
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/file"
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
	Version     string
	ServiceName string
	RepoOwner   string
	RepoName    string

	Cmd    domain.CommandRunner
	FileIO domain.FileIO
}

func NewGenerator(version, name, repoOwner, repoName string) domain.PackageGenerator {
	return &Generator{
		Version:     version,
		ServiceName: name,
		RepoOwner:   repoOwner,
		RepoName:    repoName,
		Cmd:         commandRunner.NewCommandRunner(),
		FileIO:      file.NewFileIO(),
	}
}

func (g *Generator) GeneratePackage(specificationPath, outputDir string) (string, error) {
	// Make the directory that will contain the final package
	packageDir, err := g.FileIO.MkdirAll(filepath.Join(outputDir, g.GetPackageName()), 0755)
	if err != nil {
		return "", err
	}

	// Generate package
	err = g.Cmd.ExecuteAndLog(outputDir, "npx", "openapi-generator-cli", "generate",
		"-i", specificationPath, "-g", "typescript-angular", "-o", packageDir,
		"--additional-properties=fileNaming=camelCase,ngVersion=10.0.0,stringEnums=true",
		"--enable-post-process-file", "--remove-operation-id-prefix")
	if err != nil {
		return "", errors.Wrap(err, "failed to generate package")
	}

	// Copy the package.json file & replace the version
	packageJSONPath, err := g.getPackageJSON(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to get package.json")
	}

	err = g.FileIO.CopyManyToDir(packageDir, TSConfigPath, ConfigurationTSPath)
	if err != nil {
		return "", err
	}

	err = g.installNPMPackages(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to install npm packages")
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
	err = g.FileIO.CopyManyToDir(distDir, NPMRCPath, ConfigurationTSPath, packageJSONPath)
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
