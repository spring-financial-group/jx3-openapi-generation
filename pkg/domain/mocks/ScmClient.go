// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	github "github.com/google/go-github/v47/github"

	mock "github.com/stretchr/testify/mock"
)

// ScmClient is an autogenerated mock type for the ScmClient type
type ScmClient struct {
	mock.Mock
}

// AddLabels provides a mock function with given fields: ctx, labels, pullNumber
func (_m *ScmClient) AddLabels(ctx context.Context, labels []string, pullNumber int) ([]*github.Label, error) {
	ret := _m.Called(ctx, labels, pullNumber)

	var r0 []*github.Label
	if rf, ok := ret.Get(0).(func(context.Context, []string, int) []*github.Label); ok {
		r0 = rf(ctx, labels, pullNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, int) error); ok {
		r1 = rf(ctx, labels, pullNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatePullRequest provides a mock function with given fields: ctx, newPullRequest
func (_m *ScmClient) CreatePullRequest(ctx context.Context, newPullRequest *github.NewPullRequest) (*github.PullRequest, error) {
	ret := _m.Called(ctx, newPullRequest)

	var r0 *github.PullRequest
	if rf, ok := ret.Get(0).(func(context.Context, *github.NewPullRequest) *github.PullRequest); ok {
		r0 = rf(ctx, newPullRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.PullRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *github.NewPullRequest) error); ok {
		r1 = rf(ctx, newPullRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestReviewers provides a mock function with given fields: ctx, reviewers, pullNumber
func (_m *ScmClient) RequestReviewers(ctx context.Context, reviewers []string, pullNumber int) (*github.PullRequest, error) {
	ret := _m.Called(ctx, reviewers, pullNumber)

	var r0 *github.PullRequest
	if rf, ok := ret.Get(0).(func(context.Context, []string, int) *github.PullRequest); ok {
		r0 = rf(ctx, reviewers, pullNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.PullRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, int) error); ok {
		r1 = rf(ctx, reviewers, pullNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewScmClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewScmClient creates a new instance of ScmClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewScmClient(t mockConstructorTestingTNewScmClient) *ScmClient {
	mock := &ScmClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
