// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
)

type issueService struct {
	client *wrapper
}

func (s *issueService) AssignIssue(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	panic("implement me")
}

func (s *issueService) UnassignIssue(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	panic("implement me")
}

func (s *issueService) ListEvents(context.Context, string, int, scm.ListOptions) ([]*scm.ListedIssueEvent, *scm.Response, error) {
	panic("implement me")
}

func (s *issueService) ListLabels(context.Context, string, int, scm.ListOptions) ([]*scm.Label, *scm.Response, error) {
	// TODO implement this
	return nil, nil, nil
}

func (s *issueService) AddLabel(ctx context.Context, repo string, number int, label string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *issueService) DeleteLabel(ctx context.Context, repo string, number int, label string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *issueService) Find(ctx context.Context, repo string, number int) (*scm.Issue, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *issueService) FindComment(ctx context.Context, repo string, index, id int) (*scm.Comment, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *issueService) List(ctx context.Context, repo string, opts scm.IssueListOptions) ([]*scm.Issue, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *issueService) ListComments(ctx context.Context, repo string, index int, opts scm.ListOptions) ([]*scm.Comment, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *issueService) Create(ctx context.Context, repo string, input *scm.IssueInput) (*scm.Issue, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *issueService) CreateComment(ctx context.Context, repo string, number int, in *scm.CommentInput) (*scm.Comment, *scm.Response, error) {
	input := pullRequestCommentInput{Text: in.Body}
	namespace, name := scm.Split(repo)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/issues/%d/comments", namespace, name, number)
	out := new(pullRequestComment)
	res, err := s.client.do(ctx, "POST", path, &input, out)
	return convertPullRequestComment(out), res, err
}

func (s *issueService) DeleteComment(ctx context.Context, repo string, number, id int) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *issueService) Close(ctx context.Context, repo string, number int) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *issueService) Lock(ctx context.Context, repo string, number int) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *issueService) Unlock(ctx context.Context, repo string, number int) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}
