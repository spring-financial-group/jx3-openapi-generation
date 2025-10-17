package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/file"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/rootcmd"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"
)

// Options for triggering
type Options struct {
	Apps []string
	Args []string
	Cmd  *cobra.Command

	Version            string
	SwaggerServiceName string
	RepoOwner          string
	RepoName           string
	SpecPath           string
	GitUser            string
	GitToken           string

	FileIO      domain.FileIO
	PackageName string
}

// Constants for environment variables required by the command
const (
	versionKey            = "VERSION"
	repoOwnerKey          = "REPO_OWNER"
	repoNameKey           = "REPO_NAME"
	swaggerServiceNameKey = "SwaggerServiceName"
	specPathKey           = "SpecPath"
	gitUserKey            = "GIT_USER"
	gitTokenKey           = "GIT_TOKEN"
	packageNameKey        = "PackageName"
)

const (
	validResources = `Valid resource types include:
	* packages
	`
)

var (
	generateLong = templates.LongDesc(`
		Display one or more resources.
		` + validResources + `
`)

	genExample = templates.Examples(`
		%s generate packages
	`)
)

// NewCmdGenerate creates a command object for the generic "generate" action, which
// creates on or more resources.
func NewCmdGenerate() *cobra.Command {
	o := &Options{
		FileIO: file.NewFileIO(),
	}

	// Initialising the generic variables for this command and all sub-commands
	err := o.initialise()
	helper.CheckErr(err)

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generates one or more resources",
		Long:    generateLong,
		Example: fmt.Sprintf(genExample, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			o.Cmd = cmd
			o.Args = args
			err := o.Run()
			helper.CheckErr(err)
		},
		SuggestFor: []string{"genarate, genorate"},
		Aliases:    []string{"gen"},
	}

	cmd.AddCommand(NewCmdGeneratePackages(o))
	return cmd
}

// Run implements this command
func (o *Options) Run() error {
	return o.Cmd.Help()
}

func (o *Options) initialise() error {
	err := o.getVariablesFromEnvironment()
	if err != nil {
		return err
	}
	err = o.validateSpecificationLocation()
	if err != nil {
		return err
	}
	return nil
}

func (o *Options) getVariablesFromEnvironment() error {
	var missingVariables []string
	if o.Version = os.Getenv(versionKey); o.Version == "" {
		missingVariables = append(missingVariables, versionKey)
	}
	if o.RepoOwner = os.Getenv(repoOwnerKey); o.RepoOwner == "" {
		missingVariables = append(missingVariables, repoOwnerKey)
	}
	if o.RepoName = os.Getenv(repoNameKey); o.RepoName == "" {
		missingVariables = append(missingVariables, repoNameKey)
	}
	if o.SwaggerServiceName = os.Getenv(swaggerServiceNameKey); o.SwaggerServiceName == "" {
		missingVariables = append(missingVariables, swaggerServiceNameKey)
	}
	if o.PackageName = os.Getenv(packageNameKey); o.PackageName == "" {
		o.PackageName = "Client"
	}
	if o.SpecPath = os.Getenv(specPathKey); o.SpecPath == "" {
		missingVariables = append(missingVariables, specPathKey)
	}
	if o.GitUser = os.Getenv(gitUserKey); o.GitUser == "" {
		missingVariables = append(missingVariables, gitUserKey)
	}
	if o.GitToken = os.Getenv(gitTokenKey); o.GitToken == "" {
		missingVariables = append(missingVariables, gitTokenKey)
	}
	if len(missingVariables) > 0 {
		return &domain.EnvironmentVariableNotFoundError{VariableNames: missingVariables}
	}
	return nil
}

func (o *Options) validateSpecificationLocation() error {
	absPath, err := o.getAbsoluteSpecPath(o.SpecPath)
	if err != nil {
		return errors.Wrap(err, "failed to get absolute path for specification")
	}

	exists, err := o.FileIO.Exists(absPath)
	if err != nil {
		return errors.Wrap(err, "failed to check if specification exists")
	}
	if !exists {
		return errors.Wrap(&domain.FileNotFoundError{FilePath: o.SpecPath}, "failed to check if specification exists")
	}
	o.SpecPath = absPath
	log.Info().Msgf("%sSpecification found at %s%s", utils.Cyan, absPath, utils.Reset)
	return nil
}

func (o *Options) getAbsoluteSpecPath(relativePath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, relativePath), nil
}
