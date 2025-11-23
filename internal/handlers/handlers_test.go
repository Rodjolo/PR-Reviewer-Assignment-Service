package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

type mockPRService2 struct{}

func (m *mockPRService2) CreatePR(title string, authorID int) (*models.PR, error) { return nil, nil }
func (m *mockPRService2) GetPR(id int) (*models.PR, error)                        { return nil, nil }
func (m *mockPRService2) GetAllPRs() ([]models.PR, error)                         { return nil, nil }
func (m *mockPRService2) GetPRsByUserID(userID int) ([]models.PR, error)          { return nil, nil }
func (m *mockPRService2) ReassignReviewer(prID, oldReviewerID int) (*models.PR, error) {
	return nil, nil
}
func (m *mockPRService2) MergePR(id int) (*models.PR, error) { return nil, nil }

type mockUserService2 struct{}

func (m *mockUserService2) CreateUser(name string, isActive bool) (*models.User, error) {
	return nil, nil
}
func (m *mockUserService2) GetUser(id int) (*models.User, error) { return nil, nil }
func (m *mockUserService2) GetAllUsers() ([]models.User, error)  { return nil, nil }
func (m *mockUserService2) UpdateUser(id int, name *string, isActive *bool) (*models.User, error) {
	return nil, nil
}
func (m *mockUserService2) BulkDeactivateTeam(teamName string) (*dto.BulkDeactivateTeamResponse, error) {
	return nil, nil
}

type mockTeamService2 struct{}

func (m *mockTeamService2) CreateTeam(name string) (*models.Team, error) { return nil, nil }
func (m *mockTeamService2) GetTeam(name string) (*models.Team, error)    { return nil, nil }
func (m *mockTeamService2) GetAllTeams() ([]models.Team, error)          { return nil, nil }
func (m *mockTeamService2) AddMember(teamName string, userID int) (*models.Team, error) {
	return nil, nil
}
func (m *mockTeamService2) RemoveMember(teamName string, userID int) error { return nil }

type mockStatsService2 struct{}

func (m *mockStatsService2) GetStats() (*dto.StatsResponse, error) {
	return &dto.StatsResponse{}, nil
}

func TestRespondJSON2(t *testing.T) {
	handler := NewHandlers(&mockPRService2{}, &mockUserService2{}, &mockTeamService2{}, &mockStatsService2{})

	rec := httptest.NewRecorder()
	data := map[string]string{"test": "value"}

	handler.respondJSON(rec, http.StatusOK, data)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRespondError2(t *testing.T) {
	handler := NewHandlers(&mockPRService2{}, &mockUserService2{}, &mockTeamService2{}, &mockStatsService2{})

	rec := httptest.NewRecorder()

	handler.respondError(rec, http.StatusBadRequest, "test error")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
