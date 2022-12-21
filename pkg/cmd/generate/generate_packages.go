package generate

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"spring-financial-group/jx3-openapi-generation/pkg/utils"
	"strings"
)

const (
	// OpenAPIToolsPath is the path to the openapitools.json file
	OpenAPIToolsPath = "/openapitools.json"
)

// PackageOptions contains the common options for the command
type PackageOptions struct {
	*Options

	GeneratorFactory domain.PackageGeneratorFactory
	CmdRunner        domain.CommandRunner
}

var (
	formatLong = templates.LongDesc(`
		Generates client packages from an OpenAPI/Swagger specification.
`)

	formatExample = templates.Examples(`
		# Generates client packages
		%s package java
	`)
)

// NewCmdGeneratePackages creates a command object for the generate action which generates one or more
// packages from an OpenAPI specification
func NewCmdGeneratePackages(opts *Options) *cobra.Command {
	o := &PackageOptions{
		Options: opts,
		GeneratorFactory: packageGenerator.NewFactory(
			opts.Version,
			opts.SwaggerServiceName,
			opts.RepoOwner,
			opts.RepoName,
			opts.GitToken,
		),
		CmdRunner: commandRunner.NewCommandRunner(),
	}

	cmd := &cobra.Command{
		Use:     "package",
		Short:   "generates client packages",
		Long:    formatLong,
		Example: formatExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.Cmd = cmd
			o.Args = args
			err := o.Run(args)
			helper.CheckErr(err)
		},
		SuggestFor: []string{"p", "pack", "packa", "packag"},
		Aliases:    []string{"pkg", "pkgs", "packages", "package"},
	}
	return cmd
}

// Run implements this command
func (o *PackageOptions) Run(languages []string) error {
	tmpDir, err := o.SetupEnvironment()
	if err != nil {
		return errors.Wrap(err, "failed to setup environment")
	}
	defer o.FileIO.DeferRemove(tmpDir)

	for _, l := range languages {
		log.Logger().Infof("%sGenerating %s client package%s", utils.Green, l, utils.Reset)
		gen, err := o.GeneratorFactory.NewGenerator(l)
		if err != nil {
			return errors.Wrapf(err, "failed to create generator for %s", l)
		}

		outputDir, err := o.FileIO.MkdirAll(filepath.Join(tmpDir, l), 0700)
		if err != nil {
			return errors.Wrapf(err, "failed to make output dir for %s", l)
		}

		packageDir, err := gen.GeneratePackage(o.SpecPath, outputDir)
		if err != nil {
			return errors.Wrapf(err, "failed to generate %s package", l)
		}

		log.Logger().Infof("%sPushing %s package%s", utils.Green, l, utils.Reset)
		err = gen.PushPackage(packageDir)
		if err != nil {
			return errors.Wrapf(err, "failed to push %s package", l)
		}
	}

	log.Logger().Infof("%sSuccessfully generated and pushed packages for languages: %s%s", utils.Green, strings.Join(languages, ", "), utils.Reset)
	return nil
}

// SetupEnvironment creates the output directory and copies the required files into it
func (o *PackageOptions) SetupEnvironment() (string, error) {
	log.Logger().Infof("%sSetting up environment%s", utils.Green, utils.Reset)
	tmpDir, err := o.FileIO.MkTmpDir("package-generator")
	if err != nil {
		return "", errors.Wrap(err, "failed to make tmp dir")
	}

	_, err = o.FileIO.CopyToWorkingDir(OpenAPIToolsPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to copy openapitools.json")
	}
	return tmpDir, nil
}
