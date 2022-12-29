package openapitools

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type Config struct {
	Schema       string       `json:"$schema"`
	Spaces       int          `json:"spaces"`
	GeneratorCLI GeneratorCLI `json:"generator-cli"`
}

type GeneratorCLI struct {
	Version    string               `json:"version"`
	Generators map[string]Generator `json:"generators"`
}

type Generator struct {
	Name                 string                        `json:"generatorName"`
	AdditionalProperties map[string]AdditionalProperty `json:"additionalProperties"`
}

// AdditionalProperty is a type alias for string as it can be either a string or a boolean, but we don't care about the value,
// so we can just unmarshal it into a string
type AdditionalProperty string

func (a *AdditionalProperty) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return json.Unmarshal(data, (*string)(a))
	}
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	*a = AdditionalProperty(strconv.FormatBool(b))
	return nil
}

func GetConfig(path string) (*Config, error) {
	cfg := new(Config)
	err := cfg.ReadFromFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config from %s", path)
	}
	return cfg, nil
}

func (c *Config) ReadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to read openapitools.json")
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}
	return nil
}
