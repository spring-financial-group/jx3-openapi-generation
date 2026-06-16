package swagfilter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/rootcmd"
	swagfiltercore "github.com/spring-financial-group/jx3-openapi-generation/pkg/swagfilter"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"
)

type Options struct {
	Args     []string
	Cmd      *cobra.Command
	StripTag string
	Input    string
	Output   string
}

var swagfilterLong = templates.LongDesc(`
	Strip a named tag from every operation in a Swagger 2.0 JSON specification.
	Reads the input file, removes all occurrences of the tag from operation tags
	arrays, and writes the result to the output file.
`)

var swagfilterExample = templates.Examples(`
	# Strip the "external" tag and overwrite in place
	%s swagfilter --strip-tag=external --input=docs/swagger.json

	# Strip the "external" tag and write to a new file
	%s swagfilter --strip-tag=external --input=docs/swagger.json --output=docs/swagger-filtered.json
`)

func NewCmdSwagFilter() *cobra.Command {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "swagfilter",
		Short:   "Strip a tag from all operations in a Swagger 2.0 JSON spec",
		Long:    swagfilterLong,
		Example: fmt.Sprintf(swagfilterExample, rootcmd.BinaryName, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			o.Cmd = cmd
			o.Args = args
			err := o.Run()
			helper.CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&o.StripTag, "strip-tag", "", "tag name to strip from all operation tags arrays (required)")
	cmd.Flags().StringVar(&o.Input, "input", "docs/swagger.json", "path to input swagger.json")
	cmd.Flags().StringVar(&o.Output, "output", "", "path to output file (defaults to overwriting input)")
	_ = cmd.MarkFlagRequired("strip-tag")

	return cmd
}

func (o *Options) Run() error {
	output := o.Output
	if output == "" {
		output = o.Input
	}
	output = filepath.Clean(output)

	input := filepath.Clean(o.Input)

	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", input, err)
	}

	result, err := swagfiltercore.StripTagFromSpec(data, o.StripTag)
	if err != nil {
		return fmt.Errorf("error processing swagger: %w", err)
	}

	// #nosec G703 - output path is user-controlled via CLI flag, which is intentional for a CLI tool
	if err := os.WriteFile(output, result, 0o600); err != nil {
		return fmt.Errorf("error writing %s: %w", output, err)
	}

	return nil
}
