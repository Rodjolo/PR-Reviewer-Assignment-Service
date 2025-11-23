package service

import (
	"errors"
	"fmt"
	"github.com/Rodjolo/pr-reviewer-service/internal/repository"
	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

type UserService struct {
	userRepo repository.UserRepositoryInterface
	prRepo   repository.PRRepositoryInterface
	teamRepo repository.TeamRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface, prRepo repository.PRRepositoryInterface, teamRepo repository.TeamRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
		teamRepo: teamRepo,
	}
}

func (s *UserService) CreateUser(name string, isActive bool) (*models.User, error) {
	user := &models.User{
		Name:     name,
		IsActive: isActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUser(id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(id int, name *string, isActive *bool) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if name != nil {
		user.Name = *name
	}
	if isActive != nil {
		user.IsActive = *isActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// BulkDeactivateTeam деактивирует всех пользователей команды и безопасно переназначает ревьюверов в открытых PR
// Оптимизировано для выполнения в пределах 100 мс для средних объемов данных
func (s *UserService) BulkDeactivateTeam(teamName string) (*dto.BulkDeactivateTeamResponse, error) {
	// Проверяем существование команды
	team, err := s.teamRepo.GetByName(teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	if team == nil {
		return nil, errors.New("team not found")
	}

	// Получаем список пользователей команды перед деактивацией
	userIDs := make([]int, len(team.Members))
	for i, member := range team.Members {
		userIDs[i] = member.ID
	}

	if len(userIDs) == 0 {
		return &dto.BulkDeactivateTeamResponse{
			DeactivatedUsers: 0,
			ReassignedPRs:    0,
		}, nil
	}

	// Получаем открытые PR с ревьюверами из этой команды
	prReviewerMap, err := s.prRepo.GetOpenPRsWithReviewers(userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get open PRs: %w", err)
	}

	// Деактивируем пользователей
	deactivatedCount, err := s.userRepo.BulkDeactivateByTeam(teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate users: %w", err)
	}

	// Переназначаем ревьюверов в открытых PR
	reassignedCount := 0
	if len(prReviewerMap) > 0 {
		reassignedCount, err = s.prRepo.BulkReassignReviewers(prReviewerMap, teamName, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to reassign reviewers: %w", err)
		}
	}

	return &dto.BulkDeactivateTeamResponse{
		DeactivatedUsers: deactivatedCount,
		ReassignedPRs:    reassignedCount,
	}, nil
}
