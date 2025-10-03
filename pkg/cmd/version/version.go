package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/rootcmd"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/helper"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras/templates"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
)

// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion string
)

// Options for triggering
type Options struct {
	Apps []string
	Args []string
	Cmd  *cobra.Command
}

const (
	// TestVersion used in test cases for the current version if no
	// version can be found - such as if the version property is not properly
	// included in the go test flags
	TestVersion = "1.0.0-SNAPSHOT"
)

var (
	createLong = templates.LongDesc(`
		Shows the version of mqa
`)

	createExample = templates.Examples(`
		version
	`)
)

// NewCmdVersion shows the version of the jx3-openapi-generation
func NewCmdVersion() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Shows the version of the jx3-openapi-generation",
		Long:    createLong,
		Example: fmt.Sprintf(createExample, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			o.Cmd = cmd
			o.Args = args
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.Cmd = cmd

	return cmd, o
}

func (o *Options) Run() error {
	log.Logger().Info(GetVersion())
	return nil
}

func GetVersion() string {
	if Version != "" {
		return Version
	}
	return TestVersion
}
