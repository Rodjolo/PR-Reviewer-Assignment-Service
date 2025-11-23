package service

import (
	"testing"

	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

type mockStatsPRRepository struct{}

func (m *mockStatsPRRepository) Create(pr *models.PR) error                        { return nil }
func (m *mockStatsPRRepository) GetByID(id int) (*models.PR, error)                { return nil, nil }
func (m *mockStatsPRRepository) GetByUserID(userID int) ([]models.PR, error)       { return nil, nil }
func (m *mockStatsPRRepository) GetAll() ([]models.PR, error)                      { return nil, nil }
func (m *mockStatsPRRepository) UpdateStatus(id int, status models.PRStatus) error { return nil }
func (m *mockStatsPRRepository) ReassignReviewer(prID, oldID, newID int) error     { return nil }
func (m *mockStatsPRRepository) GetOpenPRsWithReviewers(userIDs []int) (map[int][]int, error) {
	return nil, nil
}
func (m *mockStatsPRRepository) BulkReassignReviewers(prMap map[int][]int, team string, excludeIDs []int) (int, error) {
	return 0, nil
}
func (m *mockStatsPRRepository) GetStats() (map[string]int, error) {
	return map[string]int{
		"total_users":  10,
		"active_users": 8,
		"total_teams":  3,
		"total_prs":    50,
		"open_prs":     15,
		"merged_prs":   35,
	}, nil
}

func TestGetStats_Success(t *testing.T) {
	mockStats := &mockStatsPRRepository{}

	service := NewStatsService(mockStats)
	stats, err := service.GetStats()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if stats.TotalUsers != 10 {
		t.Errorf("expected 10 total users, got %d", stats.TotalUsers)
	}
	if stats.TotalTeams != 3 {
		t.Errorf("expected 3 total teams, got %d", stats.TotalTeams)
	}
	if stats.TotalPRs != 50 {
		t.Errorf("expected 50 total PRs, got %d", stats.TotalPRs)
	}
	if stats.OpenPRs != 15 {
		t.Errorf("expected 15 open PRs, got %d", stats.OpenPRs)
	}
	if stats.MergedPRs != 35 {
		t.Errorf("expected 35 merged PRs, got %d", stats.MergedPRs)
	}
	if stats.ActiveUsers != 8 {
		t.Errorf("expected 8 active users, got %d", stats.ActiveUsers)
	}
}
