package _go

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/git"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"spring-financial-group/jx3-openapi-generation/pkg/scmClient/github"
	"spring-financial-group/jx3-openapi-generation/pkg/utils"
	"strings"
)

const (
	PushRepositoryURL  = "https://github.com/spring-financial-group/mqube-go-packages.git"
	PushRepositoryName = "mqube-go-packages"

	updateBotLabel = "updatebot"
)

var (
	reviewers = []string{"Skisocks"}
)

type Generator struct {
	*packageGenerator.BaseGenerator
	Git domain.Gitter
	Scm domain.ScmClient
}

func NewGenerator(baseGenerator *packageGenerator.BaseGenerator) *Generator {
	return &Generator{
		BaseGenerator: baseGenerator,
		Git:           git.NewClient(),
		Scm:           github.NewClient(baseGenerator.RepoOwner, PushRepositoryName, baseGenerator.GitToken),
	}
}

func (g *Generator) GeneratePackage(outputDir string) (string, error) {
	repoDir, err := g.Git.Clone(outputDir, PushRepositoryURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to clone pipeline schemas")
	}

	branchName := fmt.Sprintf("update/%s/%s", g.GetPackageName(), g.Version)
	err = g.Git.CheckoutBranch(repoDir, branchName)
	if err != nil {
		return "", errors.Wrap(err, "failed to checkout branch")
	}

	packageDir, err := g.BaseGenerator.GeneratePackage(filepath.Join(repoDir, g.GetPackageName()), domain.Go)
	if err != nil {
		return "", err
	}

	err = g.Git.AddFiles(repoDir, packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to add files to Git")
	}

	err = g.Git.Commit(repoDir, fmt.Sprintf("chore(deps): upgrade %s package -> %s", g.GetPackageName(), g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to commit package")
	}
	return packageDir, nil
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

	// Add Reviewers & auto-merge labels
	_, err = g.Scm.RequestReviewers(context.Background(), reviewers, pr.GetNumber())
	if err != nil {
		return errors.Wrap(err, "failed to add reviewers to pull request")
	}
	//_, err = g.Scm.AddLabels(context.Background(), []string{updateBotLabel}, pr.GetNumber())
	//if err != nil {
	//	return errors.Wrap(err, "failed to add labels pull request")
	//}
	return nil
}
