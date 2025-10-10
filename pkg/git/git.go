package git

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/commandrunner"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/domain"
	"github.com/spring-financial-group/jx3-openapi-generation/pkg/utils"
)

type Client struct {
	cmd domain.CommandRunner
}

func NewClient() *Client {
	return &Client{
		cmd: commandrunner.NewCommandRunner(),
	}
}

func (c *Client) Clone(dir, repositoryURL string) (string, error) {
	url, err := url.Parse(repositoryURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse repository URL")
	}

	out, err := c.git(dir, "clone", repositoryURL)
	c.log(out)
	return filepath.Join(dir, strings.TrimSuffix(filepath.Base(url.Path), ".git")), err
}

func (c *Client) GetCurrentBranch(dir string) (string, error) {
	out, err := c.git(dir, "rev-parse", "--abbrev-ref", "HEAD")
	c.log(out)
	return out, err
}

func (c *Client) SetRemote(dir, repositoryURL string) error {
	_, err := c.git(dir, "remote", "set-url", "origin", repositoryURL)
	return err
}

func (c *Client) CheckoutBranch(dir, branchName string) error {
	out, err := c.git(dir, "checkout", "-b", branchName)
	c.log(out)
	return err
}

func (c *Client) AddFiles(dir string, paths ...string) error {
	out, err := c.git(dir, append([]string{"add"}, paths...)...)
	c.log(out)
	return err
}

func (c *Client) Commit(dir, message string) error {
	out, err := c.git(dir, "commit", "-m", message)
	c.log(out)
	return err
}

func (c *Client) Push(dir, branch string) error {
	out, err := c.git(dir, "push", "--set-upstream", "origin", branch, "-ff")
	c.log(out)
	return err
}

func (c *Client) GetDefaultBranchName(dir string) (string, error) {
	out, err := c.git(dir, "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	c.log(out)
	return out, err
}

func (c *Client) log(message string) {
	if message != "" {
		log.Info().Msg(message)
	}
}

func (c *Client) git(dir string, args ...string) (string, error) {
	log.Info().Msgf("%sRunning command:%s git %s", utils.Cyan, utils.Reset, strings.Join(args, " "))
	out, err := c.cmd.Execute(dir, "git", args...)
	if err != nil {
		return out, errors.Wrap(err, "failed to run git command")
	}
	return out, nil
}
