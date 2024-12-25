package database

import "time"

// Organization represents a GitHub organization in the database
type Organization struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Team represents a GitHub team in the database
type Team struct {
	ID             int64     `db:"id"`
	OrganizationID int64     `db:"organization_id"`
	Name           string    `db:"name"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
