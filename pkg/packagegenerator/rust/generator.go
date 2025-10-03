package rust

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gh "github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/git"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packagegenerator"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/scmClient/github"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
)

const (
	PushRepositoryURL  = "https://github.com/spring-financial-group/mqube-rust-packages.git"
	PushRepositoryName = "mqube-rust-packages"
	updateBotLabel     = "updatebot"
)

type Generator struct {
	*packagegenerator.BaseGenerator
	Git domain.Gitter
	Scm domain.ScmClient
}

func NewGenerator(baseGenerator *packagegenerator.BaseGenerator) *Generator {
	return &Generator{
		BaseGenerator: baseGenerator,
		Git:           git.NewClient(),
		Scm:           github.NewClient(baseGenerator.RepoOwner, PushRepositoryName, baseGenerator.GitToken),
	}
}

func (g *Generator) GeneratePackage(outputDir string) (string, error) {
	g.setDynamicConfigVariables()

	repoDir, err := g.Git.Clone(outputDir, PushRepositoryURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to clone packages repository")
	}

	branchName := fmt.Sprintf("update/%s/%s", g.GetPackageName(), g.Version)
	err = g.Git.CheckoutBranch(repoDir, branchName)
	if err != nil {
		return "", errors.Wrap(err, "failed to checkout branch")
	}

	packageDir := filepath.Join(repoDir, g.GetPackageName())
	err = g.createFreshDir(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to create fresh package dir")
	}

	_, err = g.BaseGenerator.GeneratePackage(packageDir, domain.Rust)
	if err != nil {
		return "", err
	}

	err = g.FileIO.Write(filepath.Join(packageDir, "VERSION"), []byte(g.Version), 0700)
	if err != nil {
		return "", errors.Wrap(err, "failed to write VERSION file")
	}

	err = g.Git.AddFiles(repoDir, packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to add files to Git")
	}

	err = g.Git.Commit(repoDir, fmt.Sprintf("chore(deps): upgrade %s module -> %s", g.GetPackageName(), g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to commit package")
	}

	return packageDir, nil
}

func (g *Generator) setDynamicConfigVariables() {
	g.Cfg.GeneratorCLI.Generators[domain.Rust].AdditionalProperties["packageName"] = g.GetPackageName()
	g.Cfg.GeneratorCLI.Generators[domain.Rust].AdditionalProperties["packageVersion"] = g.Version
}

func (g *Generator) GetPackageName() string {
	return g.RepoName
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

func (g *Generator) createFreshDir(packageDir string) error {
	// Check if directory exists
	if _, err := os.Stat(packageDir); err == nil {
		// Remove entire directory and its contents
		if err := os.RemoveAll(packageDir); err != nil {
			return errors.Wrapf(err, "failed to remove existing directory: %s", packageDir)
		}
		log.Logger().Infof("Removed existing directory: %s", packageDir)
	}

	// Create a fresh directory
	if err := os.MkdirAll(packageDir, 0750); err != nil {
		return errors.Wrapf(err, "failed to create directory: %s", packageDir)
	}
	log.Logger().Info("Created directory:", packageDir)

	return nil
}

func (g *Generator) createPullRequest(currentBranch, defaultBranch string) error {
	pr, err := g.Scm.CreatePullRequest(
		context.Background(),
		&gh.NewPullRequest{
			Title:               utils.NewPtr(fmt.Sprintf("chore(deps): upgrade %s package -> %s", g.GetPackageName(), g.Version)),
			Head:                &currentBranch,
			Base:                utils.NewPtr(strings.TrimPrefix(defaultBranch, "origin/")),
			Body:                utils.NewPtr(fmt.Sprintf("Automated rust package update for %s", g.GetPackageName())),
			MaintainerCanModify: utils.NewPtr(true),
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to create pull request")
	}

	// auto-merge labels
	_, err = g.Scm.AddLabels(context.Background(), []string{updateBotLabel}, pr.GetNumber())
	if err != nil {
		return errors.Wrap(err, "failed to add labels pull request")
	}
	return nil
}
