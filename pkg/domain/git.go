package domain

import (
	"context"
	"github.com/google/go-github/v47/github"
)

type ScmClient interface {
	// CreatePullRequest creates a new pull request
	CreatePullRequest(ctx context.Context, newPullRequest *github.NewPullRequest) (*github.PullRequest, error)
	// RequestReviewers requests a review on a pull request given the pr number
	RequestReviewers(ctx context.Context, reviewers []string, pullNumber int) (*github.PullRequest, error)
	// AddLabels adds labels to a pull request given the pr number
	AddLabels(ctx context.Context, labels []string, pullNumber int) ([]*github.Label, error)
}

type Gitter interface {
	// Clone clones a repo to the local env given the repo url and the directory to clone to
	Clone(dir, repositoryURL string) (string, error)
	// GetCurrentBranch gets the current branch name from the local env
	GetCurrentBranch(dir string) (string, error)
	// SetRemote sets the remote url of the local env
	SetRemote(dir, repositoryURL string) error
	// CheckoutBranch checks out a branch from the local env
	CheckoutBranch(dir, branchName string) error
	// AddFiles adds files to the local env
	AddFiles(dir string, paths ...string) error
	// Commit commits changes to the local env
	Commit(dir, message string) error
	// Push pushes changes to the remote env
	Push(dir, branch string) error
	// GetDefaultBranchName gets the default branch name of the repo from the local env
	GetDefaultBranchName(dir string) (string, error)
}
