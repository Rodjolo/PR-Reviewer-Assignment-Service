package models

// User represents a user in the system.
type User struct {
	Name     string `json:"name" db:"name"`
	ID       int    `json:"id" db:"id"`
	IsActive bool   `json:"is_active" db:"is_active"`
}
