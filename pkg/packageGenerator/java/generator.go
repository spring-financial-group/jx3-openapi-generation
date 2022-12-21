package java

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/file"
	"strings"
)

const (
	GradlePath = "./registry/build.gradle"
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
	packageDir, err := g.FileIO.MkdirAll(filepath.Join(outputDir, g.GetPackageName()), 0755)
	if err != nil {
		return "", err
	}

	// Generate Package
	err = g.Cmd.ExecuteAndLog(outputDir, "npx", "openapi-generator-cli", "generate",
		"-i", specificationPath, "-g", "java", "-o", packageDir, "--git-user-id", g.RepoOwner, "--git-repo-id", g.RepoName,
		"--global-property", "models,modelTests=false,modelDocs=false",
		fmt.Sprintf("-p basePackage=%s -p modelPackage=%s.models", g.GetPackageName(), g.GetPackageName()),
		"-p", "dateLibrary=java8-localdatetime")
	if err != nil {
		return "", errors.Wrap(err, "failed to generate package")
	}

	err = g.getBuildGradle(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to get build.gradle")
	}

	return packageDir, nil
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
