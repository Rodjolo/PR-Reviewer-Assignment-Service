// Package handlers provides HTTP request handlers for the PR reviewer service.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Rodjolo/pr-reviewer-service/internal/service"
)

type Handlers struct {
	prService    service.PRServiceInterface
	userService  service.UserServiceInterface
	teamService  service.TeamServiceInterface
	statsService service.StatsServiceInterface
}

func NewHandlers(prService service.PRServiceInterface, userService service.UserServiceInterface, teamService service.TeamServiceInterface, statsService service.StatsServiceInterface) *Handlers {
	return &Handlers{
		prService:    prService,
		userService:  userService,
		teamService:  teamService,
		statsService: statsService,
	}
}

func (h *Handlers) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but don't change response since headers are already sent
		_ = err
	}
}

func (h *Handlers) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
