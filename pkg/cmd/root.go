package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/cmd/generate"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/cmd/version"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/rootcmd"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
)

// Main creates the new command
func Main() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootcmd.TopLevelCommand,
		Short: "a CLI template",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Logger().Error(err.Error())
			}
		},
	}
	cmd.AddCommand(generate.NewCmdGenerate())
	cmd.AddCommand(cobras.SplitCommand(version.NewCmdVersion()))
	return cmd
}
