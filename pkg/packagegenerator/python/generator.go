package python

import (
	"context"
	"encoding/json"
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

	uvIndexName = "pyx"
)

var (
	reviewers = []string{"Reton2"}
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

	// For now we ignore the packageDir since this is purely for POC
	// err := g.GeneratePyxPackage(outputDir)
	// if err != nil {
	// 	return "", err
	// }

	packageDir, err := g.GenerateSchemasPackage(outputDir)
	if err != nil {
		return "", err
	}
	return packageDir, nil
}

func (g *Generator) GeneratePyxPackage(outputDir string) error {
	pyxDir, err := g.FileIO.MkdirAll(filepath.Join(outputDir, g.GetPackageName()), 0700)
	if err != nil {
		return errors.Wrap(err, "failed to create package directory")
	}

	packageDir, err := g.BaseGenerator.GeneratePackage(pyxDir, domain.Python)
	if err != nil {
		return err
	}

	err = g.Uvc.GeneratePyProjectFile(pyxDir, g.GetPackageName(), g.Version)
	if err != nil {
		return errors.Wrap(err, "failed to create pyproject.toml file")
	}

	err = g.Uvc.BuildProject(packageDir)
	if err != nil {
		return errors.Wrap(err, "failed to build UV project")
	}

	// Because this is running in parallel with schemas repo, for now the publish step will have to live in here
	// When we move to pyx we can move this to the publish step of the pipeline
	err = g.Uvc.PublishProject(packageDir, uvIndexName)
	if err != nil {
		return errors.Wrap(err, "failed to publish UV project")
	}
	log.Info().Msgf("Published UV project from %s to index %s", packageDir, uvIndexName)

	return nil
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

	packageDir, err := g.BaseGenerator.GeneratePackage(repoDir, domain.Python)
	if err != nil {
		return "", err
	}

	packageJSONPath, err := g.updatePackagesJSON(repoDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to update packages.json")
	}

	readmePath := fmt.Sprintf("%s_README.md", g.GetPackageName())
	err = g.Git.AddFiles(repoDir, packageJSONPath, g.GetPackageName(), readmePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to add package to Git")
	}

	err = g.Git.Commit(repoDir, fmt.Sprintf("chore(deps): upgrade %s package -> %s", g.GetPackageName(), g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to commit package")
	}

	return packageDir, nil
}

func (g *Generator) setDynamicConfigVariables() {
	g.Cfg.GeneratorCLI.Generators[domain.Python].AdditionalProperties["packageName"] = g.GetPackageName()
}

func (g *Generator) updatePackagesJSON(repoDir string) (string, error) {
	packagesPath := filepath.Join(repoDir, "packages.json")
	packages, err := g.getPackages(packagesPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to get packages")
	}
	newPackage := PackageInfo{
		Directory: g.GetPackageName(),
		Name:      g.RepoName,
		Version:   g.Version,
	}
	packages[newPackage.Name] = newPackage
	data, err := utils.MarshalJSON(packages)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal packages")
	}

	err = g.FileIO.Write(filepath.Join(repoDir, "packages.json"), data, 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to write packages.json")
	}
	return packagesPath, nil
}

func (g *Generator) getPackages(packagesPath string) (map[string]PackageInfo, error) {
	data, err := g.FileIO.Read(packagesPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read packages.json")
	}
	var packages map[string]PackageInfo
	err = json.Unmarshal(data, &packages)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal packages.json")
	}
	return packages, nil
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

	// Add Reviewers & auto-merge labels
	_, err = g.Scm.RequestReviewers(context.Background(), reviewers, pr.GetNumber())
	if err != nil {
		return errors.Wrap(err, "failed to add reviewers to pull request")
	}
	_, err = g.Scm.AddLabels(context.Background(), []string{updateBotLabel}, pr.GetNumber())
	if err != nil {
		return errors.Wrap(err, "failed to add labels pull request")
	}
	return nil
}
