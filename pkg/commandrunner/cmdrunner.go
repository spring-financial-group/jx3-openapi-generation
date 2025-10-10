package commandrunner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
)

type CommandRunner struct{}

func NewCommandRunner() domain.CommandRunner {
	return &CommandRunner{}
}

func (c *CommandRunner) Execute(dir, name string, args ...string) (string, error) {
	e := exec.CommandContext(context.Background(), name, args...)
	e.Dir = dir
	out, err := e.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return output, err
	}
	return output, nil
}

func (c *CommandRunner) ExecuteAndLog(dir, name string, args ...string) error {
	var dirString string
	if dir != "" {
		dirString = fmt.Sprintf(" in %s", dir)
	}
	log.Info().Msgf("%sRunning command%s:%s %s %s", utils.Cyan, dirString, utils.Reset, name, strings.Join(args, " "))
	out, err := c.Execute(dir, name, args...)
	if err != nil {
		log.Error().Msg(out)
		return err
	}
	if out != "" {
		log.Info().Msg(out)
	}
	return nil
}
