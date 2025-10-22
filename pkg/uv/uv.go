package uv

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/commandrunner"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/file"
)

type UVClient struct {
	cmd    domain.CommandRunner
	FileIO domain.FileIO
}

func NewClient() domain.UVClient {
	return &UVClient{
		cmd:    commandrunner.NewCommandRunner(),
		FileIO: file.NewFileIO(),
	}
}

func (c *UVClient) GeneratePyProjectFile(dir, packageName, packageVersion string) error {
	pkgVersion := packageVersion

	// Because:
	// https://packaging.python.org/en/latest/specifications/version-specifiers/#public-version-identifiers
	//
	// Example: 0.0.0-PR-123-12-SNAPSHOT
	if strings.Contains(pkgVersion, "SNAPSHOT") {
		pattern := `-([0-9][0-9]?[0-9]?)` // Matches any dash followed by up to 3 digits
		regMatch := regexp.MustCompile(pattern)
		matches := regMatch.FindAllStringSubmatch(pkgVersion, -1)

		var suffix string
		if len(matches) >= 2 {
			suffix = fmt.Sprintf(".preview%s.dev%s", matches[0][1], matches[1][1])
		} else {
			suffix = ".dev"
		}

		// This part replaces everything after the first dash with the suffix above
		versionPattern := `\.[0-9][0-9]?[0-9]?(-)` // Matches a dot followed by up to 3 digits and a dash
		versionRegexp := regexp.MustCompile(versionPattern)
		versionIndex := versionRegexp.FindStringIndex(pkgVersion)
		if len(versionIndex) >= 2 {
			pkgVersion = pkgVersion[:versionIndex[0]] + suffix
		} else {
			pkgVersion += suffix
		}

		log.Info().Msgf("Converted SNAPSHOT version %s to %s for pyx", packageVersion, pkgVersion)
	}

	pyProjectContent := fmt.Sprintf(`[project]
name = "%s"
version = "%s"
description = "%s schema package generated from OpenAPI specification"
readme = "%s_README.md"
classifiers = [
	"Programming Language :: Python :: 3.10",
	"Private :: pyx :: mqube"
]
requires-python = ">=3.10.0, <3.11"

[build-system]
requires = ["uv_build>=0.9.3,<0.10.0"]
build-backend = "uv_build"

[[tool.uv.index]]
name = "pyx"
url = "https://api.pyx.dev/simple/mqube/main"
publish-url = "https://api.pyx.dev/v1/upload/mqube/main"

[tool.uv.build-backend]
module-name = "%s"
module-root = ""
`, packageName, pkgVersion, packageName, packageName, packageName)

	pyProjectPath := filepath.Join(dir, "pyproject.toml")
	err := c.FileIO.Write(pyProjectPath, []byte(pyProjectContent), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write pyproject.toml")
	}
	return nil
}

func (c *UVClient) BuildProject(dir string) error {
	err := c.uvCommand(dir, "build")
	if err != nil {
		return errors.Wrap(err, "failed to build project")
	}
	return nil
}

func (c *UVClient) PublishProject(dir string, indexName string) error {
	err := c.uvCommand(dir, "publish", "--index", indexName)
	if err != nil {
		return errors.Wrap(err, "failed to publish project")
	}
	return nil
}

func (c *UVClient) uvCommand(dir string, args ...string) error {
	log.Info().Msgf("Running uv command: uv %s", args)
	out, err := c.cmd.Execute(dir, "uv", args...)
	if err != nil {
		log.Error().Msgf("uv command failed: %s", out)
		return errors.Wrap(err, "uv command failed")
	}
	return nil
}
