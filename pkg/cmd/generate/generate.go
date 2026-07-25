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
	SkipPush           bool
	ServerVariables    string

	FileIO      domain.FileIO
	PackageName string
}

// Constants for environment variables required by the command
const (
	versionKey            = "VERSION"
	repoOwnerKey          = "REPO_OWNER"
	repoNameKey           = "REPO_NAME"
	swaggerServiceNameKey = "SwaggerServiceName"
	serverVariables       = "ServerVariables"
	specPathKey           = "SpecPath"
	gitUserKey            = "GIT_USER"
	gitTokenKey           = "GIT_TOKEN"
	packageNameKey        = "PackageName"
	skipPushKey           = "SKIP_PUSH"
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			o.Cmd = cmd
			return o.validateOptions()
		},
	}

	cmd.PersistentFlags().StringVar(&o.Version, "version", os.Getenv(versionKey), fmt.Sprintf("Package version (env %s)", versionKey))
	cmd.PersistentFlags().StringVar(&o.RepoOwner, "repo-owner", os.Getenv(repoOwnerKey), fmt.Sprintf("Repository owner (env %s)", repoOwnerKey))
	cmd.PersistentFlags().StringVar(&o.RepoName, "repo-name", os.Getenv(repoNameKey), fmt.Sprintf("Repository name (env %s)", repoNameKey))
	cmd.PersistentFlags().StringVar(&o.SwaggerServiceName, "swagger-service-name", os.Getenv(swaggerServiceNameKey), fmt.Sprintf("Service name for package generation (env %s)", swaggerServiceNameKey))
	cmd.PersistentFlags().StringVar(&o.ServerVariables, "server-variables", os.Getenv(serverVariables), fmt.Sprintf("Server variables (env %s)", serverVariables))
	cmd.PersistentFlags().StringVar(&o.SpecPath, "spec-path", os.Getenv(specPathKey), fmt.Sprintf("Path to OpenAPI spec file (env %s)", specPathKey))
	cmd.PersistentFlags().StringVar(&o.GitUser, "git-user", os.Getenv(gitUserKey), fmt.Sprintf("Git username (env %s)", gitUserKey))
	cmd.PersistentFlags().StringVar(&o.GitToken, "git-token", os.Getenv(gitTokenKey), fmt.Sprintf("Git authentication token (env %s)", gitTokenKey))
	cmd.PersistentFlags().StringVar(&o.PackageName, "package-name", utils.FirstNonEmpty(os.Getenv(packageNameKey), "Client"), fmt.Sprintf("Package name (env %s, defaults to \"Client\")", packageNameKey))
	cmd.PersistentFlags().BoolVar(&o.SkipPush, "skip-push", os.Getenv(skipPushKey) == "true", fmt.Sprintf("Skip pushing generated packages (env %s)", skipPushKey))

	cmd.AddCommand(NewCmdGeneratePackages(o))
	return cmd
}

// Run implements this command
func (o *Options) Run() error {
	return o.Cmd.Help()
}

func (o *Options) validateOptions() error {
	if err := o.validateRequiredOptions(); err != nil {
		return err
	}
	return o.validateSpecificationLocation()
}

// validateRequiredOptions checks that options backed by a required flag/env var have been
// resolved to a non-empty value, whether set via CLI flag or the flag's env-derived default.
func (o *Options) validateRequiredOptions() error {
	var missingVariables []string

	required := map[string]string{
		specPathKey: o.SpecPath,
	}

	// we only require git credentials and repo info if we're not skipping the push step
	if !o.SkipPush {
		required[gitUserKey] = o.GitUser
		required[gitTokenKey] = o.GitToken
		required[versionKey] = o.Version
		required[repoOwnerKey] = o.RepoOwner
		required[repoNameKey] = o.RepoName
		required[swaggerServiceNameKey] = o.SwaggerServiceName
	} else {
		// If skipping push, we can provide default values to avoid requiring them since some generators still rely on
		// these for naming conventions or other logic but that is not relevant when not pushing.
		o.populateNoPushDefaults()
	}

	for envKey, value := range required {
		if value == "" {
			missingVariables = append(missingVariables, envKey)
		}
	}

	if len(missingVariables) > 0 {
		return &domain.EnvironmentVariableNotFoundError{VariableNames: missingVariables}
	}
	return nil
}

// populateNoPushDefaults sets default values for options that are required by some generators but not too relevant when
// skipping the push step.
func (o *Options) populateNoPushDefaults() {
	if o.RepoOwner == "" {
		o.RepoOwner = "test-owner"
	}
	if o.RepoName == "" {
		o.RepoName = "test-repo"
	}
	if o.Version == "" {
		o.Version = "0.0.0"
	}
	if o.SwaggerServiceName == "" {
		o.SwaggerServiceName = "TestService"
	}
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
	// If the path is already absolute, return it as-is
	if filepath.IsAbs(relativePath) {
		return relativePath, nil
	}

	// Otherwise, make it absolute relative to current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, relativePath), nil
}
