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

// Stats Response (используется map[string]int, но можно добавить структуру для типизации)
type StatsResponse map[string]int

