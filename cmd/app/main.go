package app

import (
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"spring-financial-group/jx3-openapi-generation/pkg/cmd"
	"spring-financial-group/jx3-openapi-generation/pkg/cmd/version"
)

// Run runs the command, if args are not nil they will be set on the command
func Run(args []string) error {
	log.Logger().Infof(`
___  ___                      _     _ _          _____                                     
|  \/  |                     | |   (_) |        /  ___|                                    
| .  . | _____   _____  ___  | |    _| | _____  \ '--.__      ____ _  __ _  __ _  ___ _ __
| |\/| |/ _ \ \ / / _ \/ __| | |   | | |/ / _ \  '--. \ \ /\ / / _' |/ _' |/ _' |/ _ \ '__|
| |  | | (_) \ V /  __/\__ \ | |___| |   <  __/ /\__/ /\ V  V / (_| | (_| | (_| |  __/ |
\_|  |_/\___/ \_/ \___||___/ \_____/_|_|\_\___| \____/  \_/\_/ \__,_|\__, |\__, |\___|_|
                                                                      __/ | __/ |
                                                                     |___/ |___/
@version: %s

`, version.GetVersion())

	rootCmd := cmd.Main()
	if args != nil {
		args = args[1:]
		rootCmd.SetArgs(args)
	}
	return rootCmd.Execute()
}
