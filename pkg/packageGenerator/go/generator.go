package _go

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	gh "github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/git"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/scmClient/github"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
)

const (
	PushRepositoryURL  = "https://github.com/spring-financial-group/mqube-go-packages.git"
	PushRepositoryName = "mqube-go-packages"
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

	g.setDynamicConfigVariables()

	packageDir, err := g.BaseGenerator.GeneratePackage(filepath.Join(repoDir, g.GetPackageName()), domain.Go)
	if err != nil {
		return "", err
	}

	// Openapitools sets the module name to the REPO_NAME, we need it to be the PushRepositoryName
	err = g.changeModuleName(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to change module name")
	}

	// Run go mod tidy to ensure the go.mod file doesn't have any unnecessary dependencies
	err = g.goModTidy(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to run go mod tidy")
	}

	// We need to be able to identify the version of the package from within the repository
	err = g.createPackageVersionFile(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to create package version file")
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

func (g *Generator) GetPackageName() string {
	return strings.ToLower(g.ServiceName)
}

func (g *Generator) setDynamicConfigVariables() {
	g.Cfg.GeneratorCLI.Generators[domain.Go].AdditionalProperties["packageName"] = g.GetPackageName()
}

func (g *Generator) changeModuleName(packageDir string) error {
	goModPath := filepath.Join(packageDir, "go.mod")
	bytes, err := g.FileIO.Read(goModPath)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	newModuleName := fmt.Sprintf("github.com/spring-financial-group/%s/%s", PushRepositoryName, g.GetPackageName())

	re, err := regexp.Compile(`module .*`)
	if err != nil {
		return errors.Wrap(err, "failed to compile regex")
	}

	bytes = re.ReplaceAll(bytes, []byte(fmt.Sprintf("module %s", newModuleName)))

	err = g.FileIO.Write(goModPath, bytes, 0700)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}
	return nil
}

func (g *Generator) goModTidy(dir string) error {
	return g.Cmd.ExecuteAndLog(dir, "go", "mod", "tidy")
}

func (g *Generator) createPackageVersionFile(packageDir string) error {
	return g.FileIO.Write(filepath.Join(packageDir, "VERSION"), []byte(g.Version), 0700)
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
	return nil
}
