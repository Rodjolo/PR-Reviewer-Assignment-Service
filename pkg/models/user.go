package models

type User struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	IsActive bool   `json:"is_active" db:"is_active"`
}
