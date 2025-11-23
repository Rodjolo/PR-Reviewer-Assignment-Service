package service

import (
	"errors"
	"testing"

	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

func TestCreateTeam_Success(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return nil, nil
		},
	}

	service := NewTeamService(mockTeam, &mockUserRepository{})
	team, err := service.CreateTeam("team1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if team == nil {
		t.Fatal("expected team to be created")
	}
	if team.Name != "team1" {
		t.Errorf("expected name 'team1', got %s", team.Name)
	}
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{Name: name}, nil
		},
	}

	service := NewTeamService(mockTeam, &mockUserRepository{})
	_, err := service.CreateTeam("team1")

	if !errors.Is(err, ErrTeamAlreadyExists) {
		t.Errorf("expected ErrTeamAlreadyExists, got %v", err)
	}
}

func TestGetTeam_Success(t *testing.T) {
	expectedTeam := &models.Team{
		Name: "team1",
		Members: []models.User{
			{ID: 1, Name: "User1", IsActive: true},
		},
	}
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return expectedTeam, nil
		},
	}

	service := NewTeamService(mockTeam, &mockUserRepository{})
	team, err := service.GetTeam("team1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if team != expectedTeam {
		t.Errorf("expected team %v, got %v", expectedTeam, team)
	}
}

func TestGetTeam_NotFound(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return nil, nil
		},
	}

	service := NewTeamService(mockTeam, &mockUserRepository{})
	_, err := service.GetTeam("nonexistent")

	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestAddMember_Success(t *testing.T) {
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: id, Name: "User", IsActive: true}, nil
		},
	}
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{
				Name:    name,
				Members: []models.User{},
			}, nil
		},
	}

	service := NewTeamService(mockTeam, mockUser)
	team, err := service.AddMember("team1", 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if team == nil {
		t.Fatal("expected team to be returned")
	}
}

func TestAddMember_UserNotFound(t *testing.T) {
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return nil, nil
		},
	}

	service := NewTeamService(&mockTeamRepository{}, mockUser)
	_, err := service.AddMember("team1", 1)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestAddMember_TeamNotFound(t *testing.T) {
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: id, Name: "User", IsActive: true}, nil
		},
	}
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return nil, nil
		},
	}

	service := NewTeamService(mockTeam, mockUser)
	_, err := service.AddMember("nonexistent", 1)

	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestRemoveMember_Success(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{
				Name: name,
				Members: []models.User{
					{ID: 1, Name: "User", IsActive: true},
				},
			}, nil
		},
	}

	service := NewTeamService(mockTeam, &mockUserRepository{})
	err := service.RemoveMember("team1", 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRemoveMember_TeamNotFound(t *testing.T) {
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return nil, nil
		},
	}

	service := NewTeamService(mockTeam, &mockUserRepository{})
	err := service.RemoveMember("nonexistent", 1)

	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}
