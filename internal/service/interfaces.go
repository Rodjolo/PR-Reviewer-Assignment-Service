package service

import (
	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

// PRServiceInterface определяет интерфейс для работы с Pull Requests
type PRServiceInterface interface {
	CreatePR(title string, authorID int) (*models.PR, error)
	GetPR(id int) (*models.PR, error)
	GetAllPRs() ([]models.PR, error)
	GetPRsByUserID(userID int) ([]models.PR, error)
	ReassignReviewer(prID int, oldReviewerID int) (*models.PR, error)
	MergePR(id int) (*models.PR, error)
}

// UserServiceInterface определяет интерфейс для работы с пользователями
type UserServiceInterface interface {
	CreateUser(name string, isActive bool) (*models.User, error)
	GetUser(id int) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(id int, name *string, isActive *bool) (*models.User, error)
	BulkDeactivateTeam(teamName string) (*dto.BulkDeactivateTeamResponse, error)
}

// TeamServiceInterface определяет интерфейс для работы с командами
type TeamServiceInterface interface {
	CreateTeam(name string) (*models.Team, error)
	GetTeam(name string) (*models.Team, error)
	GetAllTeams() ([]models.Team, error)
	AddMember(teamName string, userID int) (*models.Team, error)
	RemoveMember(teamName string, userID int) error
}

// StatsServiceInterface определяет интерфейс для работы со статистикой
type StatsServiceInterface interface {
	GetStats() (*dto.StatsResponse, error)
}
