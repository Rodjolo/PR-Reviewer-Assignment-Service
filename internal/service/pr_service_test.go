package service

import (
	"errors"
	"testing"

	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

type mockPRRepository struct {
	createFunc                  func(*models.PR) error
	getByIDFunc                 func(int) (*models.PR, error)
	getByUserIDFunc             func(int) ([]models.PR, error)
	getAllFunc                  func() ([]models.PR, error)
	updateStatusFunc            func(int, models.PRStatus) error
	reassignReviewerFunc        func(int, int, int) error
	getOpenPRsWithReviewersFunc func([]int) (map[int][]int, error)
	bulkReassignReviewersFunc   func(map[int][]int, string, []int) (int, error)
}

func (m *mockPRRepository) Create(pr *models.PR) error {
	if m.createFunc != nil {
		return m.createFunc(pr)
	}
	return nil
}

func (m *mockPRRepository) GetByID(id int) (*models.PR, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(id)
	}
	return nil, nil
}

func (m *mockPRRepository) GetByUserID(userID int) ([]models.PR, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(userID)
	}
	return nil, nil
}

func (m *mockPRRepository) GetAll() ([]models.PR, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc()
	}
	return nil, nil
}

func (m *mockPRRepository) UpdateStatus(id int, status models.PRStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(id, status)
	}
	return nil
}

func (m *mockPRRepository) ReassignReviewer(prID, oldReviewerID, newReviewerID int) error {
	if m.reassignReviewerFunc != nil {
		return m.reassignReviewerFunc(prID, oldReviewerID, newReviewerID)
	}
	return nil
}

func (m *mockPRRepository) GetOpenPRsWithReviewers(userIDs []int) (map[int][]int, error) {
	if m.getOpenPRsWithReviewersFunc != nil {
		return m.getOpenPRsWithReviewersFunc(userIDs)
	}
	return nil, nil
}

func (m *mockPRRepository) BulkReassignReviewers(prReviewerMap map[int][]int, teamName string, excludeUserIDs []int) (int, error) {
	if m.bulkReassignReviewersFunc != nil {
		return m.bulkReassignReviewersFunc(prReviewerMap, teamName, excludeUserIDs)
	}
	return 0, nil
}

func (m *mockPRRepository) GetStats() (map[string]int, error) {
	return map[string]int{}, nil
}

type mockUserRepository struct {
	getByIDFunc              func(int) (*models.User, error)
	bulkDeactivateByTeamFunc func(string) (int, error)
	getActiveUsersByTeamFunc func(string, int) ([]models.User, error)
}

func (m *mockUserRepository) GetByID(id int) (*models.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(id)
	}
	return nil, nil
}

func (m *mockUserRepository) Create(user *models.User) error { return nil }
func (m *mockUserRepository) GetAll() ([]models.User, error) { return nil, nil }
func (m *mockUserRepository) Update(user *models.User) error { return nil }

func (m *mockUserRepository) BulkDeactivateByTeam(teamName string) (int, error) {
	if m.bulkDeactivateByTeamFunc != nil {
		return m.bulkDeactivateByTeamFunc(teamName)
	}
	return 0, nil
}

func (m *mockUserRepository) GetActiveUsersByTeam(teamName string, excludeUserID int) ([]models.User, error) {
	if m.getActiveUsersByTeamFunc != nil {
		return m.getActiveUsersByTeamFunc(teamName, excludeUserID)
	}
	return []models.User{}, nil
}

type mockTeamRepository struct {
	getByNameFunc   func(string) (*models.Team, error)
	getUserTeamFunc func(int) (string, error)
}

func (m *mockTeamRepository) GetByName(name string) (*models.Team, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(name)
	}
	return nil, nil
}

func (m *mockTeamRepository) Create(team *models.Team) error                 { return nil }
func (m *mockTeamRepository) GetAll() ([]models.Team, error)                 { return nil, nil }
func (m *mockTeamRepository) AddMember(teamName string, userID int) error    { return nil }
func (m *mockTeamRepository) RemoveMember(teamName string, userID int) error { return nil }
func (m *mockTeamRepository) GetUserTeam(userID int) (string, error) {
	if m.getUserTeamFunc != nil {
		return m.getUserTeamFunc(userID)
	}
	return "team1", nil
}

func TestCreatePR_Success(t *testing.T) {
	mockPR := &mockPRRepository{
		createFunc: func(pr *models.PR) error {
			pr.ID = 1
			return nil
		},
	}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: id, Name: "Author", IsActive: true}, nil
		},
		getActiveUsersByTeamFunc: func(teamName string, excludeUserID int) ([]models.User, error) {
			return []models.User{
				{ID: 2, Name: "Reviewer1", IsActive: true},
				{ID: 3, Name: "Reviewer2", IsActive: true},
			}, nil
		},
	}
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{
				Name: name,
				Members: []models.User{
					{ID: 1, Name: "Author", IsActive: true},
					{ID: 2, Name: "Reviewer1", IsActive: true},
					{ID: 3, Name: "Reviewer2", IsActive: true},
				},
			}, nil
		},
	}

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.CreatePR("Test PR", 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pr == nil {
		t.Fatal("expected PR to be created")
	}
	if pr.ID != 1 {
		t.Errorf("expected PR ID 1, got %d", pr.ID)
	}
	if len(pr.Reviewers) != 2 {
		t.Errorf("expected 2 reviewers, got %d", len(pr.Reviewers))
	}
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	mockPR := &mockPRRepository{}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return nil, nil
		},
	}
	mockTeam := &mockTeamRepository{}

	service := NewPRService(mockPR, mockUser, mockTeam)
	_, err := service.CreatePR("Test PR", 1)

	if !errors.Is(err, ErrAuthorNotFound) {
		t.Errorf("expected ErrAuthorNotFound, got %v", err)
	}
}

func TestCreatePR_AuthorNotInTeam(t *testing.T) {
	mockPR := &mockPRRepository{}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: id, Name: "Author", IsActive: true}, nil
		},
	}
	mockTeam := &mockTeamRepository{
		getUserTeamFunc: func(userID int) (string, error) {
			return "", nil
		},
	}

	service := NewPRService(mockPR, mockUser, mockTeam)
	_, err := service.CreatePR("Test PR", 1)

	if !errors.Is(err, ErrAuthorNotInTeam) {
		t.Errorf("expected ErrAuthorNotInTeam, got %v", err)
	}
}

func TestCreatePR_NotEnoughReviewers(t *testing.T) {
	mockPR := &mockPRRepository{}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: id, Name: "Author", IsActive: true}, nil
		},
		getActiveUsersByTeamFunc: func(teamName string, excludeUserID int) ([]models.User, error) {
			return []models.User{
				{ID: 2, Name: "OnlyReviewer", IsActive: true},
			}, nil
		},
	}
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{
				Name: name,
				Members: []models.User{
					{ID: 1, Name: "Author", IsActive: true},
					{ID: 2, Name: "OnlyReviewer", IsActive: true},
				},
			}, nil
		},
	}

	service := NewPRService(mockPR, mockUser, mockTeam)
	_, err := service.CreatePR("Test PR", 1)

	if !errors.Is(err, ErrInsufficientReviewers) {
		t.Errorf("expected ErrInsufficientReviewers, got %v", err)
	}
}

func TestGetPR_Success(t *testing.T) {
	expectedPR := &models.PR{ID: 1, Title: "Test PR", AuthorID: 1}
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return expectedPR, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	pr, err := service.GetPR(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pr != expectedPR {
		t.Errorf("expected PR %v, got %v", expectedPR, pr)
	}
}

func TestGetPR_NotFound(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return nil, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	_, err := service.GetPR(1)

	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("expected ErrPRNotFound, got %v", err)
	}
}

func TestMergePR_Success(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return &models.PR{ID: id, Title: "Test", Status: models.PRStatusOpen}, nil
		},
		updateStatusFunc: func(id int, status models.PRStatus) error {
			return nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	pr, err := service.MergePR(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pr.Status != models.PRStatusMerged {
		t.Errorf("expected status merged, got %v", pr.Status)
	}
}

func TestMergePR_AlreadyMerged(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return &models.PR{ID: id, Title: "Test", Status: models.PRStatusMerged}, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	pr, err := service.MergePR(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pr.Status != models.PRStatusMerged {
		t.Errorf("expected status merged, got %v", pr.Status)
	}
}

func TestMergePR_NotFound(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return nil, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	_, err := service.MergePR(1)

	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("expected ErrPRNotFound, got %v", err)
	}
}

func TestReassignReviewer_Success(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return &models.PR{
				ID:        id,
				Title:     "Test",
				AuthorID:  1,
				Status:    models.PRStatusOpen,
				Reviewers: []int{2, 3},
			}, nil
		},
		reassignReviewerFunc: func(prID, oldID, newID int) error {
			return nil
		},
	}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id int) (*models.User, error) {
			return &models.User{ID: id, Name: "User", IsActive: true}, nil
		},
		getActiveUsersByTeamFunc: func(teamName string, excludeUserID int) ([]models.User, error) {
			return []models.User{
				{ID: 4, Name: "NewReviewer", IsActive: true},
				{ID: 5, Name: "AnotherReviewer", IsActive: true},
			}, nil
		},
	}
	mockTeam := &mockTeamRepository{
		getByNameFunc: func(name string) (*models.Team, error) {
			return &models.Team{
				Name: name,
				Members: []models.User{
					{ID: 1, Name: "Author", IsActive: true},
					{ID: 2, Name: "OldReviewer", IsActive: true},
					{ID: 3, Name: "Reviewer", IsActive: true},
					{ID: 4, Name: "NewReviewer", IsActive: true},
				},
			}, nil
		},
	}

	service := NewPRService(mockPR, mockUser, mockTeam)
	pr, err := service.ReassignReviewer(1, 2)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pr == nil {
		t.Fatal("expected PR to be returned")
	}
}

func TestReassignReviewer_PRNotFound(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return nil, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	_, err := service.ReassignReviewer(1, 2)

	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("expected ErrPRNotFound, got %v", err)
	}
}

func TestReassignReviewer_PRAlreadyMerged(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return &models.PR{ID: id, Status: models.PRStatusMerged}, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	_, err := service.ReassignReviewer(1, 2)

	if !errors.Is(err, ErrPRAlreadyMerged) {
		t.Errorf("expected ErrPRAlreadyMerged, got %v", err)
	}
}

func TestReassignReviewer_ReviewerNotAssigned(t *testing.T) {
	mockPR := &mockPRRepository{
		getByIDFunc: func(id int) (*models.PR, error) {
			return &models.PR{
				ID:        id,
				Status:    models.PRStatusOpen,
				Reviewers: []int{3, 4},
			}, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	_, err := service.ReassignReviewer(1, 2)

	if !errors.Is(err, ErrReviewerNotAssigned) {
		t.Errorf("expected ErrReviewerNotAssigned, got %v", err)
	}
}

func TestGetPRsByUserID_Success(t *testing.T) {
	expectedPRs := []models.PR{
		{ID: 1, Title: "PR1"},
		{ID: 2, Title: "PR2"},
	}
	mockPR := &mockPRRepository{
		getByUserIDFunc: func(id int) ([]models.PR, error) {
			return expectedPRs, nil
		},
	}

	service := NewPRService(mockPR, &mockUserRepository{}, &mockTeamRepository{})
	prs, err := service.GetPRsByUserID(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(prs) != 2 {
		t.Errorf("expected 2 PRs, got %d", len(prs))
	}
}
