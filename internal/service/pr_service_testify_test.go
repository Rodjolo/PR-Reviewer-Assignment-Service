package service

import (
	"testing"

	"github.com/Rodjolo/pr-reviewer-service/mocks"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePR_WithMockery_Success(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	author := &models.User{ID: 1, Name: "Author", IsActive: true}
	reviewers := []models.User{
		{ID: 2, Name: "Reviewer1", IsActive: true},
		{ID: 3, Name: "Reviewer2", IsActive: true},
	}

	mockUser.On("GetByID", 1).Return(author, nil)
	mockTeam.On("GetUserTeam", 1).Return("team1", nil)
	mockUser.On("GetActiveUsersByTeam", "team1", 1).Return(reviewers, nil)
	mockPR.On("Create", mock.MatchedBy(func(pr *models.PR) bool {
		return pr.Title == "Test PR" && pr.AuthorID == 1 && len(pr.Reviewers) == 2
	})).Run(func(args mock.Arguments) {
		pr := args.Get(0).(*models.PR)
		pr.ID = 1
	}).Return(nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.CreatePR("Test PR", 1)

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 1, pr.ID)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, 2, len(pr.Reviewers))

	mockUser.AssertExpectations(t)
	mockTeam.AssertExpectations(t)
	mockPR.AssertExpectations(t)
}

func TestCreatePR_WithMockery_AuthorNotFound(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	mockUser.On("GetByID", 999).Return(nil, nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.CreatePR("Test PR", 999)

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.ErrorIs(t, err, ErrAuthorNotFound)

	mockUser.AssertExpectations(t)
}

func TestCreatePR_WithMockery_InsufficientReviewers(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	author := &models.User{ID: 1, Name: "Author", IsActive: true}
	onlyOneReviewer := []models.User{
		{ID: 2, Name: "OnlyReviewer", IsActive: true},
	}

	mockUser.On("GetByID", 1).Return(author, nil)
	mockTeam.On("GetUserTeam", 1).Return("team1", nil)
	mockUser.On("GetActiveUsersByTeam", "team1", 1).Return(onlyOneReviewer, nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.CreatePR("Test PR", 1)

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.ErrorIs(t, err, ErrInsufficientReviewers)

	mockUser.AssertExpectations(t)
	mockTeam.AssertExpectations(t)
}

func TestGetPR_WithMockery_Success(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	expectedPR := &models.PR{
		ID:       1,
		Title:    "Test PR",
		AuthorID: 1,
		Status:   models.PRStatusOpen,
	}

	mockPR.On("GetByID", 1).Return(expectedPR, nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.GetPR(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedPR, pr)

	mockPR.AssertExpectations(t)
}

func TestGetPR_WithMockery_NotFound(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	mockPR.On("GetByID", 999).Return(nil, nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.GetPR(999)

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.ErrorIs(t, err, ErrPRNotFound)

	mockPR.AssertExpectations(t)
}

func TestMergePR_WithMockery_Success(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	existingPR := &models.PR{
		ID:       1,
		Title:    "Test PR",
		AuthorID: 1,
		Status:   models.PRStatusOpen,
	}

	mockPR.On("GetByID", 1).Return(existingPR, nil)
	mockPR.On("UpdateStatus", 1, models.PRStatusMerged).Return(nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.MergePR(1)

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, models.PRStatusMerged, pr.Status)

	mockPR.AssertExpectations(t)
	mockPR.AssertCalled(t, "GetByID", 1)
	mockPR.AssertCalled(t, "UpdateStatus", 1, models.PRStatusMerged)
}

func TestMergePR_WithMockery_AlreadyMerged(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	mergedPR := &models.PR{
		ID:       1,
		Title:    "Test PR",
		AuthorID: 1,
		Status:   models.PRStatusMerged,
	}

	mockPR.On("GetByID", 1).Return(mergedPR, nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.MergePR(1)

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, models.PRStatusMerged, pr.Status)

	mockPR.AssertExpectations(t)
	mockPR.AssertNotCalled(t, "UpdateStatus")
}

func TestReassignReviewer_WithMockery_Success(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	existingPR := &models.PR{
		ID:        1,
		Title:     "Test PR",
		AuthorID:  1,
		Status:    models.PRStatusOpen,
		Reviewers: []int{2, 3},
	}

	author := &models.User{ID: 1, Name: "Author", IsActive: true}
	oldReviewer := &models.User{ID: 2, Name: "OldReviewer", IsActive: true}
	newReviewers := []models.User{
		{ID: 4, Name: "NewReviewer", IsActive: true},
		{ID: 5, Name: "AnotherReviewer", IsActive: true},
	}

	mockPR.On("GetByID", 1).Return(existingPR, nil).Maybe()
	mockUser.On("GetByID", 1).Return(author, nil).Maybe()
	mockUser.On("GetByID", 2).Return(oldReviewer, nil).Maybe()
	mockTeam.On("GetUserTeam", 1).Return("team1", nil).Maybe()
	mockTeam.On("GetUserTeam", 2).Return("team1", nil).Maybe()
	mockUser.On("GetActiveUsersByTeam", "team1", 2).Return(newReviewers, nil).Maybe()
	mockPR.On("ReassignReviewer", 1, 2, mock.AnythingOfType("int")).Return(nil).Maybe()

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.ReassignReviewer(1, 2)

	assert.NoError(t, err)
	assert.NotNil(t, pr)

	mockPR.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockTeam.AssertExpectations(t)
}

func TestGetPRsByUserID_WithMockery_Success(t *testing.T) {
	mockPR := mocks.NewMockPRRepositoryInterface(t)
	mockUser := mocks.NewMockUserRepositoryInterface(t)
	mockTeam := mocks.NewMockTeamRepositoryInterface(t)

	expectedPRs := []models.PR{
		{ID: 1, Title: "PR1", AuthorID: 1},
		{ID: 2, Title: "PR2", AuthorID: 1},
	}

	mockPR.On("GetByUserID", 1).Return(expectedPRs, nil)

	service := NewPRService(mockPR, mockUser, mockTeam)
	prs, err := service.GetPRsByUserID(1)

	assert.NoError(t, err)
	assert.Len(t, prs, 2)
	assert.Equal(t, expectedPRs, prs)

	mockPR.AssertExpectations(t)
}
