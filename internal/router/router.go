// Package router provides HTTP routing configuration for the application.
package router

import (
	"github.com/Rodjolo/pr-reviewer-service/internal/handlers"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *handlers.Handlers) *mux.Router {
	r := mux.NewRouter()

	// PR routes
	r.HandleFunc("/prs", h.CreatePR).Methods("POST")
	r.HandleFunc("/prs", h.ListPRs).Methods("GET")
	r.HandleFunc("/prs/{id}", h.GetPR).Methods("GET")
	r.HandleFunc("/prs/{id}/reassign", h.ReassignReviewer).Methods("PATCH")
	r.HandleFunc("/prs/{id}/merge", h.MergePR).Methods("POST")

	// User routes
	r.HandleFunc("/users", h.CreateUser).Methods("POST")
	r.HandleFunc("/users", h.ListUsers).Methods("GET")
	r.HandleFunc("/users/{id}", h.GetUser).Methods("GET")
	r.HandleFunc("/users/{id}", h.UpdateUser).Methods("PATCH")

	// Team routes
	r.HandleFunc("/teams", h.CreateTeam).Methods("POST")
	r.HandleFunc("/teams", h.ListTeams).Methods("GET")
	r.HandleFunc("/teams/{name}", h.GetTeam).Methods("GET")
	r.HandleFunc("/teams/{name}/members", h.AddTeamMember).Methods("POST")
	r.HandleFunc("/teams/{name}/members", h.RemoveTeamMember).Methods("DELETE")
	r.HandleFunc("/teams/{name}/deactivate", h.BulkDeactivateTeam).Methods("POST")

	// Stats route
	r.HandleFunc("/stats", h.GetStats).Methods("GET")

	// Swagger documentation
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}
