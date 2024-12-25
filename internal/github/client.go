package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(token string) *Client {
	ctx := context.Background()
	// Token is already decrypted at config level
	client := github.NewClient(nil).WithAuthToken(token)

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

// Get team members
func (c *Client) ListTeamMembers(org, team string) ([]*github.User, error) {
	// First get the team ID
	teams, err := c.ListTeams(org)
	if err != nil {
		return nil, fmt.Errorf("error listing teams: %v", err)
	}

	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams found in organization %s", org)
	}

	// get Organization ID from one of the teams
	orgID := teams[0].GetOrganization().GetID()

	var teamID int64
	for _, t := range teams {
		if t.GetName() == team {
			teamID = t.GetID()
			break
		}
	}

	if teamID == 0 {
		return nil, fmt.Errorf("team %s not found in organization %s", team, org)
	}

	opts := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	var allMembers []*github.User
	for {
		members, resp, err := c.client.Teams.ListTeamMembersByID(c.ctx, orgID, teamID, opts)
		if err != nil {
			return nil, fmt.Errorf("error listing team members: %v", err)
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allMembers, nil
}