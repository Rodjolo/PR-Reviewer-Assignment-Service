// Package dto provides data transfer objects for API requests and responses.
package dto

// PR Requests

// CreatePRRequest represents the request body for creating a new Pull Request.
type CreatePRRequest struct {
	Title    string `json:"title" example:"Add new feature"`
	AuthorID int    `json:"author_id" example:"1"`
}

// ReassignRequest represents the request body for reassigning a PR reviewer.
type ReassignRequest struct {
	OldReviewerID int `json:"old_reviewer_id" example:"2"`
}

// User Requests

// CreateUserRequest represents the request body for creating a new user.
type CreateUserRequest struct {
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
	Name     string `json:"name" example:"Alice"`
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" example:"Alice Updated"`
	IsActive *bool   `json:"is_active,omitempty" example:"false"`
}

// Team Requests

// CreateTeamRequest represents the request body for creating a new team.
type CreateTeamRequest struct {
	Name string `json:"name" example:"backend"`
}

// AddMemberRequest represents the request body for adding a member to a team.
type AddMemberRequest struct {
	UserID int `json:"user_id" example:"1"`
}
