package dto

// Error Response

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Success Response

// MessageResponse represents a success message response from the API.
type MessageResponse struct {
	Message string `json:"message"`
}

// Bulk Deactivate Response

// BulkDeactivateTeamResponse represents the response when bulk deactivating a team.
type BulkDeactivateTeamResponse struct {
	DeactivatedUsers int `json:"deactivated_users"`
	ReassignedPRs    int `json:"reassigned_prs"`
}

// StatsResponse represents statistics about users, teams, and pull requests.
type StatsResponse struct {
	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`
	TotalTeams  int `json:"total_teams"`
	TotalPRs    int `json:"total_prs"`
	OpenPRs     int `json:"open_prs"`
	MergedPRs   int `json:"merged_prs"`
}
