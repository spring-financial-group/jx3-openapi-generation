package generate

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/commandRunner"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/openapitools"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/angular"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/csharp"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/java"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/javascript"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/python"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"path/filepath"
	_go "github.com/spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/go"
	"strings"
)

const (
	// OpenAPIToolsPath is the path to the openapitools.json file
	OpenAPIToolsPath = "/openapitools.json"
)

// PackageOptions contains the common options for the command
type PackageOptions struct {
	*Options

	languageGenerators map[string]domain.PackageGenerator
	CmdRunner          domain.CommandRunner
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
		Options:   opts,
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

	if err := o.InitialiseGenerators(); err != nil {
		helper.CheckErr(errors.Wrap(err, "failed to initialise generators"))
		return nil
	}
	if err := o.ValidateLanguages(o.Args); err != nil {
		helper.CheckErr(errors.Wrap(err, "failed to validate languages"))
		return nil
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
		outputDir, err := o.FileIO.MkdirAll(filepath.Join(tmpDir, l), 0700)
		if err != nil {
			return errors.Wrapf(err, "failed to make output dir for %s", l)
		}

		packageDir, err := o.languageGenerators[l].GeneratePackage(outputDir)
		if err != nil {
			return errors.Wrapf(err, "failed to generate %s package", l)
		}

		log.Logger().Infof("%sPushing %s package%s", utils.Green, l, utils.Reset)
		err = o.languageGenerators[l].PushPackage(packageDir)
		if err != nil {
			return errors.Wrapf(err, "failed to push %s package", l)
		}
	}

	log.Logger().Infof("%sSuccessfully generated and pushed packages for languages: %s%s", utils.Green, strings.Join(languages, ", "), utils.Reset)
	return nil
}

func (o *PackageOptions) ValidateLanguages(languages []string) error {
	for _, l := range languages {
		if _, ok := o.languageGenerators[l]; !ok {
			return &domain.ErrUnsupportedLanguage{Language: l}
		}
	}
	return nil
}

func (o *PackageOptions) InitialiseGenerators() error {
	// Get the config & init base generator
	config, err := openapitools.GetConfig()
	if err != nil {
		return err
	}
	baseGenerator, err := packageGenerator.NewBaseGenerator(o.Version, o.SwaggerServiceName, o.RepoOwner, o.RepoName, o.GitToken, o.GitUser, o.SpecPath, o.PackageName, config)
	if err != nil {
		return errors.Wrap(err, "failed to create base generator")
	}

	o.languageGenerators = map[string]domain.PackageGenerator{
		domain.CSharp:     csharp.NewGenerator(baseGenerator),
		domain.Java:       java.NewGenerator(baseGenerator),
		domain.Angular:    angular.NewGenerator(baseGenerator),
		domain.Python:     python.NewGenerator(baseGenerator),
		domain.Javascript: javscript.NewGenerator(baseGenerator),
		domain.Go:         _go.NewGenerator(baseGenerator),
	}
	return nil
}

// SetupEnvironment creates the output directory and copies the required files into it
func (o *PackageOptions) SetupEnvironment() (string, error) {
	tmpDir, err := o.FileIO.MkTmpDir("package-generator")
	if err != nil {
		return "", errors.Wrap(err, "failed to make tmp dir")
	}
	return tmpDir, nil
}
