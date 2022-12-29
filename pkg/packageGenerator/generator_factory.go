package packageGenerator

import (
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/file"
	"spring-financial-group/jx3-openapi-generation/pkg/openapitools"
	"spring-financial-group/jx3-openapi-generation/pkg/utils"
)

type BaseGenerator struct {
	Version     string
	ServiceName string
	RepoOwner   string
	RepoName    string
	GitToken    string
	SpecPath    string

	Cfg    *openapitools.Config
	Cmd    domain.CommandRunner
	FileIO domain.FileIO
}

func NewBaseGenerator(version, serviceName, repoOwner, repoName, gitToken, specPath string, cfg *openapitools.Config) (*BaseGenerator, error) {
	gen := &BaseGenerator{
		Version:     version,
		ServiceName: serviceName,
		RepoOwner:   repoOwner,
		RepoName:    repoName,
		GitToken:    gitToken,
		SpecPath:    specPath,
		Cmd:         commandRunner.NewCommandRunner(),
		FileIO:      file.NewFileIO(),
		Cfg:         cfg,
	}
	return gen, gen.initConfig()
}

// initConfig initializes the config for the generator setting the default values for the generator depending on the
// environment
func (g *BaseGenerator) initConfig() (err error) {
	for _, val := range g.Cfg.GeneratorCLI.Generators {
		val.AdditionalProperties["inputSpec"] = openapitools.AdditionalProperty(g.SpecPath)
		val.AdditionalProperties["packageVersion"] = openapitools.AdditionalProperty(g.Version)
		val.AdditionalProperties["gitRepoId"] = openapitools.AdditionalProperty(g.RepoName)
		val.AdditionalProperties["gitUserId"] = openapitools.AdditionalProperty(g.RepoOwner)
	}
	return nil
}

func (g *BaseGenerator) GeneratePackage(dir, language string) error {
	if err := g.writeConfig(dir); err != nil {
		return err
	}

	// Generate Package
	err := g.Cmd.ExecuteAndLog(dir, "npx", "openapi-generator-cli", "generate", "--generator-key", language)
	if err != nil {
		return errors.Wrap(err, "failed to generate package")
	}
	return nil
}

func (g *BaseGenerator) writeConfig(dir string) error {
	data, err := utils.MarshalJSON(g.Cfg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	// Write config to file
	if err = g.FileIO.Write(filepath.Join(dir, "openapitools.json"), data, 0755); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}
	return nil
}
