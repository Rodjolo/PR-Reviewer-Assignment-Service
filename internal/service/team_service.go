package service

import (
	"fmt"

	"github.com/Rodjolo/pr-reviewer-service/internal/repository"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

type TeamService struct {
	teamRepo repository.TeamRepositoryInterface
	userRepo repository.UserRepositoryInterface
}

func NewTeamService(teamRepo repository.TeamRepositoryInterface, userRepo repository.UserRepositoryInterface) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(name string) (*models.Team, error) {
	existing, err := s.teamRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if existing != nil {
		return nil, ErrTeamAlreadyExists
	}

	team := &models.Team{Name: name}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

func (s *TeamService) GetTeam(name string) (*models.Team, error) {
	team, err := s.teamRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if team == nil {
		return nil, ErrTeamNotFound
	}
	return team, nil
}

func (s *TeamService) GetAllTeams() ([]models.Team, error) {
	teams, err := s.teamRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return teams, nil
}

func (s *TeamService) AddMember(teamName string, userID int) (*models.Team, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	team, err := s.teamRepo.GetByName(teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if team == nil {
		return nil, ErrTeamNotFound
	}

	if err := s.teamRepo.AddMember(teamName, userID); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	updatedTeam, err := s.teamRepo.GetByName(teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated team: %w", err)
	}

	return updatedTeam, nil
}

func (s *TeamService) RemoveMember(teamName string, userID int) error {
	team, err := s.teamRepo.GetByName(teamName)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}
	if team == nil {
		return ErrTeamNotFound
	}
	if err := s.teamRepo.RemoveMember(teamName, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}
