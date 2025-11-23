package service

import (
	"errors"
	"testing"

	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

func TestCreateUser_Success(t *testing.T) {
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: 1, Name: "Test User", IsActive: true}, nil
		},
	}
	mockPR := &mockPRRepository{}
	mockTeam := &mockTeamRepository{}

	service := NewUserService(mockUser, mockPR, mockTeam)
	user, err := service.CreateUser("Test User", true)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user to be created")
	}
	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got %s", user.Name)
	}
	if !user.IsActive {
		t.Error("expected user to be active")
	}
}

func TestGetUser_Success(t *testing.T) {
	expectedUser := &models.User{ID: 1, Name: "Test User", IsActive: true}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return expectedUser, nil
		},
	}

	service := NewUserService(mockUser, &mockPRRepository{}, &mockTeamRepository{})
	user, err := service.GetUser(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user != expectedUser {
		t.Errorf("expected user %v, got %v", expectedUser, user)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return nil, nil
		},
	}

	service := NewUserService(mockUser, &mockPRRepository{}, &mockTeamRepository{})
	_, err := service.GetUser(1)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	newName := "Updated Name"
	isActive := false
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: id, Name: "Old Name", IsActive: true}, nil
		},
	}

	service := NewUserService(mockUser, &mockPRRepository{}, &mockTeamRepository{})
	user, err := service.UpdateUser(1, &newName, &isActive)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Name != newName {
		t.Errorf("expected name %s, got %s", newName, user.Name)
	}
	if user.IsActive != isActive {
		t.Errorf("expected isActive %v, got %v", isActive, user.IsActive)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	newName := "Updated Name"
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return nil, nil
		},
	}

	service := NewUserService(mockUser, &mockPRRepository{}, &mockTeamRepository{})
	_, err := service.UpdateUser(1, &newName, nil)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestBulkDeactivateTeam_Success(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{
				Name: name,
				Members: []models.User{
					{ID: 1, Name: "User1", IsActive: true},
					{ID: 2, Name: "User2", IsActive: true},
				},
			}, nil
		},
	}
	mockUser := &mockUserRepository{
		bulkDeactivateByTeamFunc: func(teamName string) (int, error) {
			return 2, nil
		},
	}
	mockPR := &mockPRRepository{
		getOpenPRsWithReviewersFunc: func(userIDs []int) (map[int][]int, error) {
			return map[int][]int{
				1: {1, 2},
			}, nil
		},
		bulkReassignReviewersFunc: func(prMap map[int][]int, team string, excludeIDs []int) (int, error) {
			return 1, nil
		},
	}

	service := NewUserService(mockUser, mockPR, mockTeam)
	response, err := service.BulkDeactivateTeam("team1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.DeactivatedUsers != 2 {
		t.Errorf("expected 2 deactivated users, got %d", response.DeactivatedUsers)
	}
	if response.ReassignedPRs != 1 {
		t.Errorf("expected 1 reassigned PR, got %d", response.ReassignedPRs)
	}
}

func TestBulkDeactivateTeam_TeamNotFound(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return nil, nil
		},
	}

	service := NewUserService(&mockUserRepository{}, &mockPRRepository{}, mockTeam)
	_, err := service.BulkDeactivateTeam("nonexistent")

	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestBulkDeactivateTeam_EmptyTeam(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{
				Name:    name,
				Members: []models.User{},
			}, nil
		},
	}

	service := NewUserService(&mockUserRepository{}, &mockPRRepository{}, mockTeam)
	response, err := service.BulkDeactivateTeam("team1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.DeactivatedUsers != 0 {
		t.Errorf("expected 0 deactivated users, got %d", response.DeactivatedUsers)
	}
	if response.ReassignedPRs != 0 {
		t.Errorf("expected 0 reassigned PRs, got %d", response.ReassignedPRs)
	}
}
