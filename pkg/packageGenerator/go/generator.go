package _go

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	gh "github.com/google/go-github/v47/github"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	reviewers = []string{"Skisocks", "Reton2"}
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

	packageDir := filepath.Join(repoDir, g.GetPackageName())

	err = g.createFreshDir(packageDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to create fresh directory")
	}

	code, err := g.generateCode()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate code")
	}

	// write code to file
	codeFile := filepath.Join(packageDir, "client_generated.go")
	err = g.FileIO.Write(codeFile, []byte(code), 0700)
	if err != nil {
		return "", errors.Wrap(err, "failed to write code to file")
	}

	// Openapitools sets the module name to the REPO_NAME, we need it to be the PushRepositoryName
	err = g.goModInit(packageDir)
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

func (g *Generator) goModInit(dir string) error {
	newModuleName := fmt.Sprintf("github.com/spring-financial-group/%s/%s", PushRepositoryName, g.GetPackageName())
	return g.Cmd.ExecuteAndLog(dir, "go", "mod", "init", newModuleName)
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
			Body:                utils.NewPtr(fmt.Sprintf("Automated go schemas update for %s", g.GetPackageName())),
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

func (g *Generator) createFreshDir(packageDir string) error {
    // Check if directory exists
    if _, err := os.Stat(packageDir); err == nil {
        // Remove entire directory and its contents
        if err := os.RemoveAll(packageDir); err != nil {
            return errors.Wrapf(err, "failed to remove existing directory: %s", packageDir)
        }
        logrus.Infof("Removed existing directory: %s", packageDir)
    }

    // Create a fresh directory
    if err := os.MkdirAll(packageDir, 0755); err != nil {
        return errors.Wrapf(err, "failed to create directory: %s", packageDir)
    }
    fmt.Println("Created directory:", packageDir)

	return nil
}

func (g *Generator) generateCode() (string, error) {
	// Read file
	var swaggerData []byte

	swaggerData, err := os.ReadFile(g.BaseGenerator.SpecPath)
	if err != nil {
		return "", err
	}

	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData(swaggerData)
	if err != nil {
		return "", errors.Wrap(err, "failed to load spec")
	}

	if strings.HasPrefix(swagger.OpenAPI, "3.1.") {
		logrus.Warnf("You are using an OpenAPI 3.1.x specification, which is not yet supported by oapi-codegen. Some functionality may not be available. Until oapi-codegen supports OpenAPI 3.1, it is recommended to downgrade your spec to 3.0.x")
	} else if strings.HasPrefix(swagger.OpenAPI, "2.") {
		swaggerData, err = g.convertSwaggerV2toV3(swaggerData)
		if err != nil {
			return "", errors.Wrap(err, "failed to convert spec to v3")
		}
	}

	swagger, err = loader.LoadFromData(swaggerData)
	if err != nil {
		return "", errors.Wrap(err, "failed to load spec")
	}

	config := codegen.Configuration{
		PackageName: g.GetPackageName(),
		Generate: codegen.GenerateOptions{
			Client:       true,
			Models:       true,
		},
	}

	code, err := codegen.Generate(swagger, config)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate code")
	}

	return code, nil
}

func (g *Generator) convertSwaggerV2toV3(data []byte) ([]byte, error) {
	var response []byte
	// Unmarshal into a map
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return response, err
	}

	// Marshal map back to JSON
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return response, err
	}

	// POST request
	resp, err := http.Post("https://converter.swagger.io/api/convert", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	// read response to str
	body := new(bytes.Buffer)
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return response, err
	}
	response = body.Bytes()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("failed to convert spec")
	}

	return response, nil
}
