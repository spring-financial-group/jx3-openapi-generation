package packageGenerator

import (
	"github.com/pkg/errors"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/file"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/openapitools"
)

type BaseGenerator struct {
	Version     string
	ServiceName string
	RepoOwner   string
	RepoName    string
	GitToken    string
	GitUser     string
	SpecPath    string
	PackageName string

	Cfg    *openapitools.Config
	Cmd    domain.CommandRunner
	FileIO domain.FileIO
}

func NewBaseGenerator(version, serviceName, repoOwner, repoName, gitToken, gitUser, specPath string, packageName string, cfg *openapitools.Config) (*BaseGenerator, error) {
	gen := &BaseGenerator{
		Version:     version,
		ServiceName: serviceName,
		RepoOwner:   repoOwner,
		RepoName:    repoName,
		GitUser:     gitUser,
		GitToken:    gitToken,
		SpecPath:    specPath,
		PackageName: packageName,
		Cmd:         commandRunner.NewCommandRunner(),
		FileIO:      file.NewFileIO(),
		Cfg:         cfg,
	}
	return gen, gen.setDynamicConfigVariables()
}

// setDynamicConfigVariables initializes the config for the generator setting the default values for the generator depending on the
// environment
func (g *BaseGenerator) setDynamicConfigVariables() (err error) {
	for _, val := range g.Cfg.GeneratorCLI.Generators {
		val.InputSpec = g.SpecPath
		val.GitRepoID = g.RepoName
		val.GitUserID = g.RepoOwner
		val.AdditionalProperties["packageVersion"] = g.Version
	}
	return nil
}

// GeneratePackage generates the package for the given language using the openapi-generator-cli. The config is written
// to the directory before running the command.
func (g *BaseGenerator) GeneratePackage(outputDir, language string) (string, error) {
	_, err := g.FileIO.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", err
	}

	// Find the generator config for this language
	var generator *openapitools.Generator
	for key, gen := range g.Cfg.GeneratorCLI.Generators {
		if key == language {
			generator = gen
			break
		}
	}

	if generator == nil {
		return "", errors.New("generator configuration not found for language: " + language)
	}

	generator.Output = outputDir
	cfgPath, err := g.Cfg.WriteToCurrentWorkingDirectory()
	if err != nil {
		return "", err
	}
	defer g.FileIO.DeferRemove(cfgPath)

	// Generate Package
	err = g.Cmd.ExecuteAndLog("", "npx", "@openapitools/openapi-generator-cli", "generate", "--generator-key", language, "--config", cfgPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate package")
	}
	return outputDir, nil
}
