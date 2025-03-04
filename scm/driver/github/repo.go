// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

type repository struct {
	ID    int `json:"id"`
	Owner struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Private       bool      `json:"private"`
	Fork          bool      `json:"fork"`
	HTMLURL       string    `json:"html_url"`
	SSHURL        string    `json:"ssh_url"`
	CloneURL      string    `json:"clone_url"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Permissions   struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	} `json:"permissions"`
}

type hook struct {
	ID     int      `json:"id,omitempty"`
	Name   string   `json:"name"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
	Config struct {
		URL         string `json:"url"`
		Secret      string `json:"secret"`
		ContentType string `json:"content_type"`
	} `json:"config"`
}

type repositoryService struct {
	client *wrapper
}

// IsCollaborator returns whether or not the user is a collaborator of the repo.
// From GitHub's API reference:
// For organization-owned repositories, the list of collaborators includes
// outside collaborators, organization members that are direct collaborators,
// organization members with access through team memberships, organization
// members with access through default organization permissions, and
// organization owners.
//
// See https://developer.github.com/v3/repos/collaborators/
func (s *repositoryService) IsCollaborator(ctx context.Context, repo, user string) (bool, *scm.Response, error) {
	req := &scm.Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("repos/%s/collaborators/%s", repo, user),
		Header: map[string][]string{
			// This accept header enables the nested teams preview.
			// https://developer.github.com/changes/2017-08-30-preview-nested-teams/
			"Accept": {"application/vnd.github.hellcat-preview+json"},
		},
	}
	res, err := s.client.doRequest(ctx, req, nil, nil)
	if err != nil {
		return false, res, err
	}
	code := res.Status
	if code == 204 {
		return true, res, nil
	} else if code == 404 {
		return false, res, nil
	}
	return false, res, fmt.Errorf("unexpected status: %d", code)
}

// ListCollaborators gets a list of all users who have access to a repo (and
// can become assignees or requested reviewers).
//
// See 'IsCollaborator' for more details.
// See https://developer.github.com/v3/repos/collaborators/
func (s *repositoryService) ListCollaborators(ctx context.Context, repo string) ([]scm.User, *scm.Response, error) {
	req := &scm.Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("repos/%s/collaborators", repo),
		Header: map[string][]string{
			// This accept header enables the nested teams preview.
			// https://developer.github.com/changes/2017-08-30-preview-nested-teams/
			"Accept": {"application/vnd.github.hellcat-preview+json"},
		},
	}
	out := []user{}
	res, err := s.client.doRequest(ctx, req, nil, &out)
	return convertUsers(out), res, err
}

// Find returns the repository by name.
func (s *repositoryService) Find(ctx context.Context, repo string) (*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out), res, err
}

// FindHook returns a repository hook.
func (s *repositoryService) FindHook(ctx context.Context, repo string, id string) (*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/hooks/%s", repo, id)
	out := new(hook)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertHook(out), res, err
}

// FindPerms returns the repository permissions.
func (s *repositoryService) FindPerms(ctx context.Context, repo string) (*scm.Perm, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out).Perm, res, err
}

// FindPerms returns the repository permissions.
//
// https://developer.github.com/v3/repos/collaborators/#review-a-users-permission-level
func (s *repositoryService) FindUserPermission(ctx context.Context, repo string, user string) (string, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/collaborators/%s/permission", repo, user)
	var out struct {
		Perm string `json:"permission"`
	}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return out.Perm, res, err
}

// List returns the user repository list.
func (s *repositoryService) List(ctx context.Context, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("user/repos?%s", encodeListOptions(opts))
	out := []*repository{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertRepositoryList(out), res, err
}

// ListHooks returns a list or repository hooks.
func (s *repositoryService) ListHooks(ctx context.Context, repo string, opts scm.ListOptions) ([]*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/hooks?%s", repo, encodeListOptions(opts))
	out := []*hook{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertHookList(out), res, err
}

// ListStatus returns a list of commit statuses.
func (s *repositoryService) ListStatus(ctx context.Context, repo, ref string, opts scm.ListOptions) ([]*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/statuses/%s?%s", repo, ref, encodeListOptions(opts))
	out := []*status{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertStatusList(out), res, err
}

// FindCombinedStatus returns the latest statuses for a given ref.
//
// See https://developer.github.com/v3/repos/statuses/#get-the-combined-status-for-a-specific-ref
func (s *repositoryService) FindCombinedStatus(ctx context.Context, repo, ref string) (*scm.CombinedStatus, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/statuses/%s", repo, ref)
	out := &combinedStatus{}
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertCombinedStatus(out), res, err
}

func (s *repositoryService) ListLabels(ctx context.Context, repo string, opts scm.ListOptions) ([]*scm.Label, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/labels?%s", repo, encodeListOptions(opts))
	out := []*label{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertLabelObjects(out), res, err
}

// CreateHook creates a new repository webhook.
func (s *repositoryService) CreateHook(ctx context.Context, repo string, input *scm.HookInput) (*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/hooks", repo)
	in := new(hook)
	in.Active = true
	in.Name = "web"
	in.Config.Secret = input.Secret
	in.Config.ContentType = "json"
	in.Config.URL = input.Target
	in.Events = append(
		input.NativeEvents,
		convertHookEvents(input.Events)...,
	)
	out := new(hook)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertHook(out), res, err
}

// CreateStatus creates a new commit status.
func (s *repositoryService) CreateStatus(ctx context.Context, repo, ref string, input *scm.StatusInput) (*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/statuses/%s", repo, ref)
	in := &status{
		State:       convertFromState(input.State),
		Context:     input.Label,
		Description: input.Desc,
		TargetURL:   input.Target,
	}
	out := new(status)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertStatus(out), res, err
}

// DeleteHook deletes a repository webhook.
func (s *repositoryService) DeleteHook(ctx context.Context, repo string, id string) (*scm.Response, error) {
	path := fmt.Sprintf("repos/%s/hooks/%s", repo, id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

// helper function to convert from the gogs repository list to
// the common repository structure.
func convertRepositoryList(from []*repository) []*scm.Repository {
	to := []*scm.Repository{}
	for _, v := range from {
		to = append(to, convertRepository(v))
	}
	return to
}

// helper function to convert from the gogs repository structure
// to the common repository structure.
func convertRepository(from *repository) *scm.Repository {
	return &scm.Repository{
		ID:        strconv.Itoa(from.ID),
		Name:      from.Name,
		Namespace: from.Owner.Login,
		FullName:  from.FullName,
		Perm: &scm.Perm{
			Push:  from.Permissions.Push,
			Pull:  from.Permissions.Pull,
			Admin: from.Permissions.Admin,
		},
		Link:     from.HTMLURL,
		Branch:   from.DefaultBranch,
		Private:  from.Private,
		Clone:    from.CloneURL,
		CloneSSH: from.SSHURL,
		Created:  from.CreatedAt,
		Updated:  from.UpdatedAt,
	}
}

func convertHookList(from []*hook) []*scm.Hook {
	to := []*scm.Hook{}
	for _, v := range from {
		to = append(to, convertHook(v))
	}
	return to
}

func convertHook(from *hook) *scm.Hook {
	return &scm.Hook{
		ID:     strconv.Itoa(from.ID),
		Active: from.Active,
		Target: from.Config.URL,
		Events: from.Events,
	}
}

func convertHookEvents(from scm.HookEvents) []string {
	var events []string
	if from.Push {
		events = append(events, "push")
	}
	if from.PullRequest {
		events = append(events, "pull_request")
	}
	if from.PullRequestComment {
		events = append(events, "pull_request_review_comment")
	}
	if from.Issue {
		events = append(events, "issues")
	}
	if from.IssueComment || from.PullRequestComment {
		events = append(events, "issue_comment")
	}
	if from.Branch || from.Tag {
		events = append(events, "create")
		events = append(events, "delete")
	}
	return events
}

type combinedStatus struct {
	Sha      string    `json:"sha"`
	Statuses []*status `json:"statuses"`
	State    string    `json:"state"`
}

type status struct {
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	State       string    `json:"state"`
	TargetURL   string    `json:"target_url"`
	Description string    `json:"description"`
	Context     string    `json:"context"`
}

func convertCombinedStatus(from *combinedStatus) *scm.CombinedStatus {
	return &scm.CombinedStatus{
		Sha:      from.Sha,
		State:    convertState(from.State),
		Statuses: convertStatusList(from.Statuses),
	}
}

func convertStatusList(from []*status) []*scm.Status {
	to := []*scm.Status{}
	for _, v := range from {
		to = append(to, convertStatus(v))
	}
	return to
}

func convertStatus(from *status) *scm.Status {
	return &scm.Status{
		State:  convertState(from.State),
		Label:  from.Context,
		Desc:   from.Description,
		Target: from.TargetURL,
	}
}

func convertState(from string) scm.State {
	switch from {
	case "error":
		return scm.StateError
	case "failure":
		return scm.StateFailure
	case "pending":
		return scm.StatePending
	case "success":
		return scm.StateSuccess
	default:
		return scm.StateUnknown
	}
}

func convertFromState(from scm.State) string {
	switch from {
	case scm.StatePending, scm.StateRunning:
		return "pending"
	case scm.StateSuccess:
		return "success"
	case scm.StateFailure:
		return "failure"
	default:
		return "error"
	}
}
