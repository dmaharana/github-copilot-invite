package database

import (
	"fmt"
	"github-copilot-invite/internal/github"
)

// SyncWithGitHub synchronizes the local database with GitHub data
func (m *Manager) SyncWithGitHub(client *github.Client) error {
	// Sync organizations
	orgs, err := client.ListOrganizations()
	if err != nil {
		return fmt.Errorf("error listing organizations: %v", err)
	}

	for _, org := range orgs {
		dbOrg := &Organization{
			ID:   org.GetID(),
			Name: org.GetLogin(),
		}
		if err := m.UpsertOrganization(dbOrg); err != nil {
			return fmt.Errorf("error upserting organization %s: %v", org.GetLogin(), err)
		}

		// Sync teams for this organization
		teams, err := client.ListTeams(org.GetLogin())
		if err != nil {
			return fmt.Errorf("error listing teams for org %s: %v", org.GetLogin(), err)
		}

		for _, team := range teams {
			dbTeam := &Team{
				ID:             team.GetID(),
				OrganizationID: org.GetID(),
				Name:           team.GetName(),
			}
			if err := m.UpsertTeam(dbTeam); err != nil {
				return fmt.Errorf("error upserting team %s: %v", team.GetName(), err)
			}
		}
	}

	return nil
}
