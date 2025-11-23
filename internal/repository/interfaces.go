package repository

import (
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

// PRRepositoryInterface определяет интерфейс для работы с Pull Requests
type PRRepositoryInterface interface {
	Create(pr *models.PR) error
	GetByID(id int) (*models.PR, error)
	GetByUserID(userID int) ([]models.PR, error)
	GetAll() ([]models.PR, error)
	UpdateStatus(id int, status models.PRStatus) error
	ReassignReviewer(prID int, oldReviewerID int, newReviewerID int) error
	GetStats() (map[string]int, error)
	GetOpenPRsWithReviewers(userIDs []int) (map[int][]int, error)
	BulkReassignReviewers(prReviewerMap map[int][]int, teamName string, excludeUserIDs []int) (int, error)
}

// UserRepositoryInterface определяет интерфейс для работы с пользователями
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id int) (*models.User, error)
	GetAll() ([]models.User, error)
	Update(user *models.User) error
	GetActiveUsersByTeam(teamName string, excludeUserID int) ([]models.User, error)
	BulkDeactivateByTeam(teamName string) (int, error)
}

// TeamRepositoryInterface определяет интерфейс для работы с командами
type TeamRepositoryInterface interface {
	Create(team *models.Team) error
	GetByName(name string) (*models.Team, error)
	GetAll() ([]models.Team, error)
	AddMember(teamName string, userID int) error
	RemoveMember(teamName string, userID int) error
	GetUserTeam(userID int) (string, error)
}

