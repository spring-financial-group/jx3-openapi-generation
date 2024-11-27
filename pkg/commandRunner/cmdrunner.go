package commandRunner

import (
	"fmt"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"os/exec"
	"strings"
)

type CommandRunner struct{}

func NewCommandRunner() domain.CommandRunner {
	return &CommandRunner{}
}

func (c *CommandRunner) Execute(dir, name string, args ...string) (string, error) {
	e := exec.Command(name, args...)
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
	log.Logger().Infof("%sRunning command%s:%s %s %s", utils.Cyan, dirString, utils.Reset, name, strings.Join(args, " "))
	out, err := c.Execute(dir, name, args...)
	if err != nil {
		log.Logger().Errorf("%d", out)
		return err
	}
	if out != "" {
		log.Logger().Infof(out)
	}
	return nil
}
