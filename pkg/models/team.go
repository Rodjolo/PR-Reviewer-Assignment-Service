package models

// Team represents a team in the system.
type Team struct {
	Name    string `json:"name" db:"name"`
	Members []User `json:"members"`
}
