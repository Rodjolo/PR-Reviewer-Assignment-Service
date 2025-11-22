package dto

// PR Requests

type CreatePRRequest struct {
	Title    string `json:"title" example:"Add new feature"`
	AuthorID int    `json:"author_id" example:"1"`
}

type ReassignRequest struct {
	OldReviewerID int `json:"old_reviewer_id" example:"2"`
}

// User Requests

type CreateUserRequest struct {
	Name     string `json:"name" example:"Alice"`
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" example:"Alice Updated"`
	IsActive *bool   `json:"is_active,omitempty" example:"false"`
}

// Team Requests

type CreateTeamRequest struct {
	Name string `json:"name" example:"backend"`
}

type AddMemberRequest struct {
	UserID int `json:"user_id" example:"1"`
}

