package models

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PR struct {
	ID        int      `json:"id" db:"id"`
	Title     string   `json:"title" db:"title"`
	AuthorID  int      `json:"author_id" db:"author_id"`
	Status    PRStatus `json:"status" db:"status"`
	Reviewers []int    `json:"reviewers" db:"reviewers"`
}
