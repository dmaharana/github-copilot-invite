package database

import (
	"encoding/json"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

var (
	organizationBucket = []byte("organizations")
	teamBucket        = []byte("teams")
)

// Manager handles database operations
type Manager struct {
	db *bbolt.DB
}

// NewManager creates a new database manager
func NewManager(dbPath string) (*Manager, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	m := &Manager{db: db}
	if err := m.createBuckets(); err != nil {
		return nil, fmt.Errorf("error creating buckets: %v", err)
	}

	return m, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// createBuckets creates the necessary database buckets if they don't exist
func (m *Manager) createBuckets() error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(organizationBucket); err != nil {
			return fmt.Errorf("error creating organizations bucket: %v", err)
		}
		if _, err := tx.CreateBucketIfNotExists(teamBucket); err != nil {
			return fmt.Errorf("error creating teams bucket: %v", err)
		}
		return nil
	})
}

// UpsertOrganization creates or updates an organization
func (m *Manager) UpsertOrganization(org *Organization) error {
	now := time.Now()
	if org.CreatedAt.IsZero() {
		org.CreatedAt = now
	}
	org.UpdatedAt = now

	return m.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(organizationBucket)
		
		data, err := json.Marshal(org)
		if err != nil {
			return fmt.Errorf("error marshaling organization: %v", err)
		}

		key := []byte(fmt.Sprintf("%d", org.ID))
		return b.Put(key, data)
	})
}

// UpsertTeam creates or updates a team
func (m *Manager) UpsertTeam(team *Team) error {
	now := time.Now()
	if team.CreatedAt.IsZero() {
		team.CreatedAt = now
	}
	team.UpdatedAt = now

	return m.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(teamBucket)
		
		data, err := json.Marshal(team)
		if err != nil {
			return fmt.Errorf("error marshaling team: %v", err)
		}

		key := []byte(fmt.Sprintf("%d:%s", team.OrganizationID, team.Name))
		return b.Put(key, data)
	})
}

// GetOrganizationByName retrieves an organization by its name
func (m *Manager) GetOrganizationByName(name string) (*Organization, error) {
	var org *Organization

	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(organizationBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var o Organization
			if err := json.Unmarshal(v, &o); err != nil {
				return fmt.Errorf("error unmarshaling organization: %v", err)
			}
			if o.Name == name {
				org = &o
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return org, nil
}

// GetTeamByName retrieves a team by its name and organization ID
func (m *Manager) GetTeamByName(orgID int64, name string) (*Team, error) {
	var team *Team

	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(teamBucket)
		key := []byte(fmt.Sprintf("%d:%s", orgID, name))
		
		data := b.Get(key)
		if data == nil {
			return nil
		}

		team = &Team{}
		return json.Unmarshal(data, team)
	})

	if err != nil {
		return nil, err
	}
	return team, nil
}

// ListOrganizations returns all organizations
func (m *Manager) ListOrganizations() ([]*Organization, error) {
	var orgs []*Organization

	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(organizationBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var org Organization
			if err := json.Unmarshal(v, &org); err != nil {
				return fmt.Errorf("error unmarshaling organization: %v", err)
			}
			orgs = append(orgs, &org)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return orgs, nil
}

// ListTeams returns all teams for an organization
func (m *Manager) ListTeams(orgID int64) ([]*Team, error) {
	var teams []*Team
	prefix := []byte(fmt.Sprintf("%d:", orgID))

	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(teamBucket)
		c := b.Cursor()

		for k, v := c.Seek(prefix); k != nil && hasPrefix(k, prefix); k, v = c.Next() {
			var team Team
			if err := json.Unmarshal(v, &team); err != nil {
				return fmt.Errorf("error unmarshaling team: %v", err)
			}
			teams = append(teams, &team)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return teams, nil
}

// hasPrefix checks if bytes b has prefix p
func hasPrefix(b, p []byte) bool {
	if len(b) < len(p) {
		return false
	}
	for i := range p {
		if b[i] != p[i] {
			return false
		}
	}
	return true
}
