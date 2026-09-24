package python

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	gh "github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/git"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packagegenerator"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/scmClient/github"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/uv"
)

const (
	PipelineSchemasURL  = "https://github.com/spring-financial-group/mqube-ml-doc-pipeline-schemas.git"
	PipelineSchemasName = "mqube-ml-doc-pipeline-schemas"

	updateBotLabel = "updatebot"
)

type Generator struct {
	*packagegenerator.BaseGenerator
	Git domain.Gitter
	Scm domain.ScmClient
	Uvc domain.UVClient
}

func NewGenerator(baseGenerator *packagegenerator.BaseGenerator) *Generator {
	return &Generator{
		BaseGenerator: baseGenerator,
		Git:           git.NewClient(),
		Scm:           github.NewClient(baseGenerator.RepoOwner, PipelineSchemasName, baseGenerator.GitToken),
		Uvc:           uv.NewClient(),
	}
}

func (g *Generator) GeneratePackage(outputDir string) (string, error) {
	g.setDynamicConfigVariables()

	packageDir, err := g.GenerateSchemasPackage(outputDir)
	if err != nil {
		return "", err
	}
	return packageDir, nil
}

func (g *Generator) GenerateSchemasPackage(outputDir string) (string, error) {
	repoDir, err := g.Git.Clone(outputDir, PipelineSchemasURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to clone pipeline schemas")
	}

	branchName := fmt.Sprintf("update/%s/%s", g.GetPackageName(), g.Version)
	err = g.Git.CheckoutBranch(repoDir, branchName)
	if err != nil {
		return "", errors.Wrap(err, "failed to checkout branch")
	}

	// Start from a clean slate so that files removed from the spec don't linger from the previous generation
	err = g.createFreshDir(filepath.Join(repoDir, g.GetPackageName()))
	if err != nil {
		return "", errors.Wrap(err, "failed to create fresh package directory")
	}

	packageDir, err := g.BaseGenerator.GeneratePackage(repoDir, domain.Python)
	if err != nil {
		return "", err
	}

	readmePath := fmt.Sprintf("%s_README.md", g.GetPackageName())
	err = g.Git.AddFiles(repoDir, g.GetPackageName(), readmePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to add package to Git")
	}

	err = g.Git.Commit(repoDir, fmt.Sprintf("chore(deps): upgrade %s package -> %s", g.GetPackageName(), g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to commit package")
	}

	return packageDir, nil
}

func (g *Generator) createFreshDir(packageDir string) error {
	exists, err := g.FileIO.Exists(packageDir)
	if err != nil {
		return errors.Wrapf(err, "failed to check if directory exists: %s", packageDir)
	}
	if exists {
		if err := g.FileIO.Remove(packageDir); err != nil {
			return errors.Wrapf(err, "failed to remove existing directory: %s", packageDir)
		}
		log.Info().Msgf("Removed existing directory: %s", packageDir)
	}

	if _, err := g.FileIO.MkdirAll(packageDir, 0750); err != nil {
		return errors.Wrapf(err, "failed to create directory: %s", packageDir)
	}
	log.Info().Msgf("Created directory: %s", packageDir)

	return nil
}

func (g *Generator) setDynamicConfigVariables() {
	g.Cfg.GeneratorCLI.Generators[domain.Python].AdditionalProperties["packageName"] = g.GetPackageName()
}

type PackageInfo struct {
	Directory string `json:"dir"`
	Name      string `json:"name"`
	Version   string `json:"version"`
}

func (g *Generator) GetPackageName() string {
	return strings.ReplaceAll(g.RepoName, "-", "_")
}

func (g *Generator) PushPackage(packageDir string) error {
	currentBranch, err := g.Git.GetCurrentBranch(packageDir)
	if err != nil {
		return errors.Wrap(err, "failed to get current branch")
	}

	err = g.Git.Push(packageDir, currentBranch)
	if err != nil {
		return errors.Wrap(err, "failed to Git push package")
	}

	defaultBranch, err := g.Git.GetDefaultBranchName(packageDir)
	if err != nil {
		return errors.Wrap(err, "failed to get default branch name")
	}

	err = g.createPullRequest(currentBranch, defaultBranch)
	if err != nil {
		return errors.Wrap(err, "failed to create pull request")
	}
	return nil
}

func (g *Generator) createPullRequest(currentBranch, defaultBranch string) error {
	pr, err := g.Scm.CreatePullRequest(
		context.Background(),
		&gh.NewPullRequest{
			Title:               utils.NewPtr(fmt.Sprintf("chore(deps): upgrade %s package -> %s", g.GetPackageName(), g.Version)),
			Head:                &currentBranch,
			Base:                utils.NewPtr(strings.TrimPrefix(defaultBranch, "origin/")),
			Body:                utils.NewPtr(fmt.Sprintf("Automated python schemas update for %s", g.GetPackageName())),
			MaintainerCanModify: utils.NewPtr(true),
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to create pull request")
	}

	// Add auto-merge label
	_, err = g.Scm.AddLabels(context.Background(), []string{updateBotLabel}, pr.GetNumber())
	if err != nil {
		return errors.Wrap(err, "failed to add labels pull request")
	}
	return nil
}
