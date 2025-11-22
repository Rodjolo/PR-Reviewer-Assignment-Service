package models

type Team struct {
	Name    string  `json:"name" db:"name"`
	Members []User  `json:"members"`
}

