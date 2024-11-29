package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func (c *Client) ListOrganizations() ([]*github.Organization, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}
	var allOrgs []*github.Organization
	for {
		orgs, resp, err := c.client.Organizations.List(c.ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("error listing organizations: %v", err)
		}
		allOrgs = append(allOrgs, orgs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allOrgs, nil
}

func (c *Client) ListTeams(org string) ([]*github.Team, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}
	var allTeams []*github.Team
	for {
		teams, resp, err := c.client.Teams.ListTeams(c.ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("error listing teams for org %s: %v", org, err)
		}
		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allTeams, nil
}

func (c *Client) CreateTeam(org string, team *github.NewTeam) (*github.Team, error) {
	newTeam, _, err := c.client.Teams.CreateTeam(c.ctx, org, *team)
	if err != nil {
		return nil, fmt.Errorf("error creating team in org %s: %v", org, err)
	}
	return newTeam, nil
}

func (c *Client) SendCopilotInvite(org, team, username string) error {
	// Note: This is a placeholder for the actual Copilot invite API
	// GitHub's API for Copilot management might require specific endpoints or permissions
	return fmt.Errorf("copilot invite functionality not implemented")
}
