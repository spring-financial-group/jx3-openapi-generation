package github

import (
	"context"
	"github.com/google/go-github/v47/github"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"golang.org/x/oauth2"
	"strings"
)

type Client struct {
	Github *github.Client

	Owner string
	Repo  string
}

func NewClient(owner, repo, token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		Github: github.NewClient(tc),
		Owner:  owner,
		Repo:   repo,
	}
}

func (c *Client) CreatePullRequest(ctx context.Context, pullRequest *github.NewPullRequest) (*github.PullRequest, error) {
	log.Logger().Infof("Creating pull request for %s/%s", c.Owner, c.Repo)
	pr, _, err := c.Github.PullRequests.Create(ctx, c.Owner, c.Repo, pullRequest)
	if err != nil {
		return nil, err
	}
	log.Logger().Infof("Pull Request created at %s", pr.GetHTMLURL())
	return pr, nil
}

func (c *Client) RequestReviewers(ctx context.Context, reviewers []string, pullNumber int) (*github.PullRequest, error) {
	log.Logger().Infof("Requesting reviewers (%s) for %s/%s-PR-%d", strings.Join(reviewers, ", "), c.Owner, c.Repo, pullNumber)
	pr, _, err := c.Github.PullRequests.RequestReviewers(ctx, c.Owner, c.Repo, pullNumber, github.ReviewersRequest{Reviewers: reviewers})
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (c *Client) AddLabels(ctx context.Context, labels []string, pullNumber int) ([]*github.Label, error) {
	log.Logger().Infof("Adding labels (%s) for %s/%s-PR-%d", strings.Join(labels, ", "), c.Owner, c.Repo, pullNumber)
	lbs, _, err := c.Github.Issues.AddLabelsToIssue(ctx, c.Owner, c.Repo, pullNumber, labels)
	if err != nil {
		return nil, err
	}
	return lbs, nil
}
