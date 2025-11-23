package service

import (
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
	"github.com/stretchr/testify/mock"
)

type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) Create(pr *models.PR) error {
	args := m.Called(pr)
	return args.Error(0)
}

func (m *MockPRRepository) GetByID(id int) (*models.PR, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PR), args.Error(1)
}

func (m *MockPRRepository) GetByUserID(userID int) ([]models.PR, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PR), args.Error(1)
}

func (m *MockPRRepository) GetAll() ([]models.PR, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PR), args.Error(1)
}

func (m *MockPRRepository) UpdateStatus(id int, status models.PRStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockPRRepository) ReassignReviewer(prID, oldReviewerID, newReviewerID int) error {
	args := m.Called(prID, oldReviewerID, newReviewerID)
	return args.Error(0)
}

func (m *MockPRRepository) GetStats() (map[string]int, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockPRRepository) GetOpenPRsWithReviewers(userIDs []int) (map[int][]int, error) {
	args := m.Called(userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int][]int), args.Error(1)
}

func (m *MockPRRepository) BulkReassignReviewers(prReviewerMap map[int][]int, teamName string, excludeUserIDs []int) (int, error) {
	args := m.Called(prReviewerMap, teamName, excludeUserIDs)
	return args.Int(0), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAll() ([]models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetActiveUsersByTeam(teamName string, excludeUserID int) ([]models.User, error) {
	args := m.Called(teamName, excludeUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) BulkDeactivateByTeam(teamName string) (int, error) {
	args := m.Called(teamName)
	return args.Int(0), args.Error(1)
}

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(team *models.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByName(name string) (*models.Team, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetAll() ([]models.Team, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamRepository) AddMember(teamName string, userID int) error {
	args := m.Called(teamName, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) RemoveMember(teamName string, userID int) error {
	args := m.Called(teamName, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) GetUserTeam(userID int) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}
