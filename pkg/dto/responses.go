package dto

// Error Response

type ErrorResponse struct {
	Error string `json:"error"`
}

// Success Response

type MessageResponse struct {
	Message string `json:"message"`
}

// Bulk Deactivate Response

type BulkDeactivateTeamResponse struct {
	DeactivatedUsers int `json:"deactivated_users"`
	ReassignedPRs    int `json:"reassigned_prs"`
}

// Stats Response
type StatsResponse struct {
	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`
	TotalTeams  int `json:"total_teams"`
	TotalPRs    int `json:"total_prs"`
	OpenPRs     int `json:"open_prs"`
	MergedPRs   int `json:"merged_prs"`
}
