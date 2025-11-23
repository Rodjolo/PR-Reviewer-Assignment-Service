// Package models defines data models for the application.
package models

// PRStatus represents the status of a pull request.
type PRStatus string

const (
	// PRStatusOpen indicates that the PR is open and awaiting review.
	PRStatusOpen PRStatus = "OPEN"
	// PRStatusMerged indicates that the PR has been merged.
	PRStatusMerged PRStatus = "MERGED"
)

// PR represents a pull request in the system.
type PR struct {
	Title     string   `json:"title" db:"title"`
	Status    PRStatus `json:"status" db:"status"`
	Reviewers []int    `json:"reviewers" db:"reviewers"`
	ID        int      `json:"id" db:"id"`
	AuthorID  int      `json:"author_id" db:"author_id"`
}
