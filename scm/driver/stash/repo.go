// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/jenkins-x/go-scm/scm"
)

type repository struct {
	Slug          string `json:"slug"`
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ScmID         string `json:"scmId"`
	State         string `json:"state"`
	StatusMessage string `json:"statusMessage"`
	Forkable      bool   `json:"forkable"`
	Project       struct {
		Key    string `json:"key"`
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Public bool   `json:"public"`
		Type   string `json:"type"`
		Links  struct {
			Self []link `json:"self"`
		} `json:"links"`
	} `json:"project"`
	Public bool `json:"public"`
	Links  struct {
		Clone []link `json:"clone"`
		Self  []link `json:"self"`
	} `json:"links"`
}

type repositories struct {
	pagination
	Values []*repository `json:"values"`
}

type link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type perms struct {
	Values []*perm `json:"values"`
}

type perm struct {
	Permissions string `json:"permission"`
}

type hooks struct {
	pagination
	Values []*hook `json:"values"`
}

type hook struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	CreatedDate int64    `json:"createdDate"`
	UpdatedDate int64    `json:"updatedDate"`
	Events      []string `json:"events"`
	URL         string   `json:"url"`
	Active      bool     `json:"active"`
	Config      struct {
		Secret string `json:"secret"`
	} `json:"configuration"`
}

type hookInput struct {
	Name   string   `json:"name"`
	Events []string `json:"events"`
	URL    string   `json:"url"`
	Active bool     `json:"active"`
	Config struct {
		Secret string `json:"secret"`
	} `json:"configuration"`
}

type status struct {
	State string `json:"state"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Desc  string `json:"description"`
}

type participants struct {
	pagination
	Values []*participant `json:"values"`
}

type participant struct {
	User       user   `json:"user"`
	Permission string `json:"permission"`
}

type repositoryService struct {
	client *wrapper
}

func (s *repositoryService) FindCombinedStatus(ctx context.Context, repo, ref string) (*scm.CombinedStatus, *scm.Response, error) {
	panic("implement me")
}

func (s *repositoryService) FindUserPermission(ctx context.Context, repo string, user string) (string, *scm.Response, error) {
	panic("implement me")
}

func (s *repositoryService) IsCollaborator(ctx context.Context, repo, user string) (bool, *scm.Response, error) {
	users, resp, err := s.ListCollaborators(ctx, repo)
	if err != nil {
		return false, resp, err
	}
	for _, u := range users {
		if u.Name == user || u.Login == user {
			return true, resp, err
		}
	}
	return false, resp, err
}

func (s *repositoryService) ListCollaborators(ctx context.Context, repo string) ([]scm.User, *scm.Response, error) {
	namespace, name := scm.Split(repo)
	opts := scm.ListOptions{
		Size: 1000,
	}
	//path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/participants?role=PARTICIPANT&%s", namespace, name, encodeListOptions(opts))
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/permissions/users?%s", namespace, name, encodeListOptions(opts))
	out := new(participants)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	if !out.pagination.LastPage.Bool {
		res.Page.First = 1
		res.Page.Next = opts.Page + 1
	}
	return convertParticipants(out), res, err
}

func (s *repositoryService) ListLabels(context.Context, string, scm.ListOptions) ([]*scm.Label, *scm.Response, error) {
	// TODO implement me!
	return nil, nil, nil
}

// Find returns the repository by name.
func (s *repositoryService) Find(ctx context.Context, repo string) (*scm.Repository, *scm.Response, error) {
	namespace, name := scm.Split(repo)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s", namespace, name)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out), res, err
}

// FindHook returns a repository hook.
func (s *repositoryService) FindHook(ctx context.Context, repo string, id string) (*scm.Hook, *scm.Response, error) {
	namespace, name := scm.Split(repo)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/webhooks/%s", namespace, name, id)
	out := new(hook)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertHook(out), res, err
}

// FindPerms returns the repository permissions.
func (s *repositoryService) FindPerms(ctx context.Context, repo string) (*scm.Perm, *scm.Response, error) {
	// HACK: test if the user has read access to the repository.
	_, _, err := s.Find(ctx, repo)
	if err != nil {
		return &scm.Perm{
			Pull:  false,
			Push:  false,
			Admin: false,
		}, nil, nil
	}

	// HACK: test if the user has admin access to the repository.
	_, _, err = s.ListHooks(ctx, repo, scm.ListOptions{})
	if err == nil {
		return &scm.Perm{
			Pull:  true,
			Push:  true,
			Admin: true,
		}, nil, nil
	}
	// HACK: test if the user has write access to the repository.
	_, name := scm.Split(repo)
	repos, _, _ := s.listWrite(ctx, repo)
	for _, repo := range repos {
		if repo.Name == name {
			return &scm.Perm{
				Pull:  true,
				Push:  true,
				Admin: false,
			}, nil, nil
		}
	}

	return &scm.Perm{
		Pull:  true,
		Push:  false,
		Admin: false,
	}, nil, nil
}

// List returns the user repository list.
func (s *repositoryService) List(ctx context.Context, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("rest/api/1.0/repos?%s", encodeListRoleOptions(opts))
	out := new(repositories)
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	if !out.pagination.LastPage.Bool {
		res.Page.First = 1
		res.Page.Next = opts.Page + 1
	}
	return convertRepositoryList(out), res, err
}

// listWrite returns the user repository list.
func (s *repositoryService) listWrite(ctx context.Context, repo string) ([]*scm.Repository, *scm.Response, error) {
	namespace, name := scm.Split(repo)
	path := fmt.Sprintf("rest/api/1.0/repos?size=1000&permission=REPO_WRITE&project=%s&name=%s", namespace, name)
	out := new(repositories)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepositoryList(out), res, err
}

// ListHooks returns a list or repository hooks.
func (s *repositoryService) ListHooks(ctx context.Context, repo string, opts scm.ListOptions) ([]*scm.Hook, *scm.Response, error) {
	namespace, name := scm.Split(repo)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/webhooks?%s", namespace, name, encodeListOptions(opts))
	out := new(hooks)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	if !out.pagination.LastPage.Bool {
		res.Page.First = 1
		res.Page.Next = opts.Page + 1
	}
	return convertHookList(out), res, err
}

// ListStatus returns a list of commit statuses.
func (s *repositoryService) ListStatus(ctx context.Context, repo, ref string, opts scm.ListOptions) ([]*scm.Status, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

// CreateHook creates a new repository webhook.
func (s *repositoryService) CreateHook(ctx context.Context, repo string, input *scm.HookInput) (*scm.Hook, *scm.Response, error) {
	namespace, name := scm.Split(repo)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/webhooks", namespace, name)
	in := new(hookInput)
	in.URL = input.Target
	in.Active = true
	in.Name = input.Name
	in.Config.Secret = input.Secret
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
	path := fmt.Sprintf("rest/build-status/1.0/commits/%s", ref)
	in := status{
		State: convertFromState(input.State),
		Key:   input.Label,
		Name:  input.Label,
		URL:   input.Target,
		Desc:  input.Desc,
	}
	res, err := s.client.do(ctx, "POST", path, in, nil)
	return &scm.Status{
		State:  input.State,
		Label:  input.Label,
		Desc:   input.Desc,
		Target: input.Target,
	}, res, err
}

// DeleteHook deletes a repository webhook.
func (s *repositoryService) DeleteHook(ctx context.Context, repo string, id string) (*scm.Response, error) {
	namespace, name := scm.Split(repo)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/webhooks/%s", namespace, name, id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

// helper function to convert from the gogs repository list to
// the common repository structure.
func convertRepositoryList(from *repositories) []*scm.Repository {
	to := []*scm.Repository{}
	for _, v := range from.Values {
		to = append(to, convertRepository(v))
	}
	return to
}

// helper function to convert from the gogs repository structure
// to the common repository structure.
func convertRepository(from *repository) *scm.Repository {
	return &scm.Repository{
		ID:        strconv.Itoa(from.ID),
		Name:      from.Slug,
		Namespace: from.Project.Key,
		Link:      extractSelfLink(from.Links.Self),
		Branch:    "master",
		Private:   !from.Public,
		CloneSSH:  extractLink(from.Links.Clone, "ssh"),
		Clone:     anonymizeLink(extractLink(from.Links.Clone, "http")),
	}
}

func extractLink(links []link, name string) (href string) {
	for _, link := range links {
		if link.Name == name {
			return link.Href
		}
	}
	return
}

func extractSelfLink(links []link) (href string) {
	for _, link := range links {
		return link.Href
	}
	return
}

func anonymizeLink(link string) (href string) {
	parsed, err := url.Parse(link)
	if err != nil {
		return link
	}
	parsed.User = nil
	return parsed.String()
}

func convertHookList(from *hooks) []*scm.Hook {
	to := []*scm.Hook{}
	for _, v := range from.Values {
		to = append(to, convertHook(v))
	}
	return to
}

func convertHook(from *hook) *scm.Hook {
	return &scm.Hook{
		ID:     strconv.Itoa(from.ID),
		Name:   from.Name,
		Active: from.Active,
		Target: from.URL,
		Events: from.Events,
	}
}

func convertHookEvents(from scm.HookEvents) []string {
	var events []string
	if from.Push || from.Branch || from.Tag {
		events = append(events, "repo:refs_changed")
	}
	if from.PullRequest {
		events = append(events, "pr:declined")
		events = append(events, "pr:modified")
		events = append(events, "pr:deleted")
		events = append(events, "pr:opened")
		events = append(events, "pr:merged")
	}
	if from.PullRequestComment {
		events = append(events, "pr:comment:added")
		events = append(events, "pr:comment:deleted")
		events = append(events, "pr:comment:edited")
	}
	return events
}

func convertFromState(from scm.State) string {
	switch from {
	case scm.StatePending, scm.StateRunning:
		return "INPROGRESS"
	case scm.StateSuccess:
		return "SUCCESSFUL"
	default:
		return "FAILED"
	}
}

func convertState(from string) scm.State {
	switch from {
	case "FAILED":
		return scm.StateFailure
	case "INPROGRESS":
		return scm.StatePending
	case "SUCCESSFUL":
		return scm.StateSuccess
	default:
		return scm.StateUnknown
	}
}

func convertParticipants(participants *participants) []scm.User {
	answer := []scm.User{}
	for _, p := range participants.Values {
		answer = append(answer, *convertUser(&p.User))
	}
	return answer
}
