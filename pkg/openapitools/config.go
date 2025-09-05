package openapitools

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
)

const (
	OpenAPIConfigFileName = "openapitools.json"
	ConfigsDir            = "/configs"
)

var (
	initialConfigPath = "/" + OpenAPIConfigFileName
)

type Config struct {
	Schema       string       `json:"$schema"`
	Spaces       int          `json:"spaces"`
	GeneratorCLI GeneratorCLI `json:"generator-cli"`
}

type GeneratorCLI struct {
	Version    string                `json:"version"`
	Generators map[string]*Generator `json:"generators"`
}

type Generator struct {
	Name                    string            `json:"generatorName"`
	Output                  string            `json:"output"`
	InputSpec               string            `json:"inputSpec"`
	GitRepoID               string            `json:"gitRepoId,omitempty"`
	GitUserID               string            `json:"gitUserId,omitempty"`
	EnablePostProcessFile   bool              `json:"enablePostProcessFile,omitempty"`
	RemoveOperationIdPrefix bool              `json:"removeOperationIdPrefix,omitempty"`
	GlobalProperty          map[string]string `json:"globalProperty,omitempty"`
	AdditionalProperties    map[string]string `json:"additionalProperties,omitempty"`
}

func GetConfig() (*Config, error) {
	cfg := new(Config)
	err := cfg.readFromFile(initialConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config from %s", initialConfigPath)
	}

	// Openapi-generator reads the config from the current working directory so when we first get the config we should
	// move it to the cwd.
	_, err = cfg.WriteToCurrentWorkingDirectory()
	if err != nil {
		return nil, errors.Wrap(err, "failed to write config to file")
	}
	return cfg, nil
}

func GetConfigForLanguage(language string) (*Config, error) {
	cfg := new(Config)
	configPath := filepath.Join(ConfigsDir, language+"-"+OpenAPIConfigFileName)
	err := cfg.readFromFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config from %s", configPath)
	}
	return cfg, nil
}

func (c *Config) readFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to read openapitools.json")
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}

	// Initialise the maps if they're nil so we can add to them
	for _, val := range c.GeneratorCLI.Generators {
		if val.GlobalProperty == nil {
			val.GlobalProperty = make(map[string]string)
		}
		if val.AdditionalProperties == nil {
			val.AdditionalProperties = make(map[string]string)
		}
	}
	return nil
}

func (c *Config) WriteToCurrentWorkingDirectory() (string, error) {
	data, err := utils.MarshalJSON(c)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal config")
	}

	path := filepath.Join("./", OpenAPIConfigFileName)
	err = os.WriteFile(path, data, 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to write config to directory")
	}
	return path, nil
}
