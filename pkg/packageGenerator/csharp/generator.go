package csharp

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/file"
)

const (
	NugetConfigPath = "./registry/nuget.config"
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
		"-i", specificationPath, "--generator-key", "csharp", "-o", packageDir, "--git-user-id", g.RepoOwner,
		"--git-repo-id", g.RepoName,
		fmt.Sprintf("--additional-properties=packageName=%s,packageVersion=%s", g.GetPackageName(), g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to generate package")
	}

	// Copy nuget.config
	_, _, err = g.FileIO.CopyToDir(NugetConfigPath, packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to copy nuget config")
	}

	err = g.Cmd.ExecuteAndLog(packageDir, "dotnet", "pack", "-c", "Release", fmt.Sprintf("-p:VERSION=%s", g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to pack solution")
	}
	return packageDir, nil
}

func (g *Generator) GetPackageName() string {
	return fmt.Sprintf("Mqube.%s.Client", g.ServiceName)
}

func (g *Generator) PushPackage(packageDir string) error {
	solutionPath := fmt.Sprintf("./src/%s/bin/Release/**/*.nupkg", g.GetPackageName())
	return g.Cmd.ExecuteAndLog(packageDir, "dotnet", "nuget", "push", solutionPath, "-s", "mqube.packages", "--skip-duplicate")
}
