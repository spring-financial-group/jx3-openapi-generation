package python

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/json"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/file"
	"spring-financial-group/jx3-openapi-generation/pkg/git"
	"spring-financial-group/jx3-openapi-generation/pkg/scmClient/github"
	"spring-financial-group/jx3-openapi-generation/pkg/utils"
	"strings"
)

const (
	PipelineSchemasURL  = "https://github.com/spring-financial-group/mqube-ml-doc-pipeline-schemas.git"
	PipelineSchemasName = "mqube-ml-doc-pipeline-schemas"

	updateBotLabel = "updatebot"
)

var (
	reviewers = []string{"Reton2", "stelios93"}
)

type Generator struct {
	Version     string
	ServiceName string
	RepoOwner   string
	RepoName    string

	Git    domain.Gitter
	Scm    domain.ScmClient
	Cmd    domain.CommandRunner
	FileIO domain.FileIO
}

func NewGenerator(version, name, repoOwner, repoName, token string) domain.PackageGenerator {
	return &Generator{
		Version:     version,
		ServiceName: name,
		RepoOwner:   repoOwner,
		RepoName:    repoName,
		Scm:         github.NewClient(repoOwner, PipelineSchemasName, token),
		Git:         git.NewClient(),
		Cmd:         commandRunner.NewCommandRunner(),
		FileIO:      file.NewFileIO(),
	}
}

func (g *Generator) GeneratePackage(specificationPath, outputDir string) (string, error) {
	generatorDir, err := g.FileIO.MkdirAll(filepath.Join(outputDir, g.GetPackageName()), 0755)
	if err != nil {
		return "", err
	}

	repoDir, err := g.Git.Clone(generatorDir, PipelineSchemasURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to clone pipeline schemas")
	}

	branchName := fmt.Sprintf("update/%s/%s", g.GetPackageName(), g.Version)
	err = g.Git.CheckoutBranch(repoDir, branchName)
	if err != nil {
		return "", errors.Wrap(err, "failed to checkout branch")
	}

	packageDir, err := g.FileIO.MkdirAll(filepath.Join(repoDir, g.GetPackageName()), 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to create package dir")
	}

	err = g.Cmd.ExecuteAndLog(packageDir, "datamodel-codegen", "--input", specificationPath,
		"--input-file-type", "auto", "--output", "schemas.py")
	if err != nil {
		return "", errors.Wrap(err, "failed to generate package")
	}

	initPy := filepath.Join(packageDir, "__init__.py")
	err = g.FileIO.Write(initPy, []byte{}, 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to create init file")
	}

	packageJsonPath, err := g.updatePackagesJSON(repoDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to update packages.json")
	}

	schemasPy := filepath.Join(packageDir, "schemas.py")
	err = g.Git.AddFiles(repoDir, schemasPy, initPy, packageJsonPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to add package to Git")
	}

	err = g.Git.Commit(repoDir, fmt.Sprintf("chore(deps): upgrade %s package -> %s", g.GetPackageName(), g.Version))
	if err != nil {
		return "", errors.Wrap(err, "failed to commit package")
	}
	return packageDir, nil
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
