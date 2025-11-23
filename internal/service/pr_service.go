package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Rodjolo/pr-reviewer-service/internal/repository"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

type PRService struct {
	prRepo   repository.PRRepositoryInterface
	userRepo repository.UserRepositoryInterface
	teamRepo repository.TeamRepositoryInterface
}

func NewPRService(prRepo repository.PRRepositoryInterface, userRepo repository.UserRepositoryInterface, teamRepo repository.TeamRepositoryInterface) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *PRService) CreatePR(title string, authorID int) (*models.PR, error) {
	author, err := s.userRepo.GetByID(authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	if author == nil {
		return nil, ErrAuthorNotFound
	}

	teamName, err := s.teamRepo.GetUserTeam(authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user team: %w", err)
	}
	if teamName == "" {
		return nil, ErrAuthorNotInTeam
	}

	candidates, err := s.userRepo.GetActiveUsersByTeam(teamName, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	if len(candidates) < 2 {
		return nil, ErrInsufficientReviewers
	}

	reviewers := s.selectRandomReviewers(candidates, 2)

	pr := &models.PR{
		Title:     title,
		AuthorID:  authorID,
		Status:    models.PRStatusOpen,
		Reviewers: reviewers,
	}

	if err := s.prRepo.Create(pr); err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr, nil
}

func (s *PRService) selectRandomReviewers(candidates []models.User, maxCount int) []int {
	if len(candidates) == 0 {
		return []int{}
	}

	count := maxCount
	if len(candidates) < maxCount {
		count = len(candidates)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shuffled := make([]models.User, len(candidates))
	copy(shuffled, candidates)
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	reviewers := make([]int, 0, count)
	for i := 0; i < count; i++ {
		reviewers = append(reviewers, shuffled[i].ID)
	}

	return reviewers
}

func (s *PRService) GetPR(id int) (*models.PR, error) {
	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}
	if pr == nil {
		return nil, ErrPRNotFound
	}
	return pr, nil
}

func (s *PRService) GetPRsByUserID(userID int) ([]models.PR, error) {
	prs, err := s.prRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs: %w", err)
	}
	return prs, nil
}

func (s *PRService) GetAllPRs() ([]models.PR, error) {
	prs, err := s.prRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs: %w", err)
	}
	return prs, nil
}

func (s *PRService) MergePR(id int) (*models.PR, error) {
	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}
	if pr == nil {
		return nil, ErrPRNotFound
	}

	if pr.Status == models.PRStatusMerged {
		return pr, nil
	}

	if err := s.prRepo.UpdateStatus(id, models.PRStatusMerged); err != nil {
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	pr.Status = models.PRStatusMerged
	return pr, nil
}

func (s *PRService) ReassignReviewer(prID int, oldReviewerID int) (*models.PR, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}
	if pr == nil {
		return nil, ErrPRNotFound
	}

	if pr.Status == models.PRStatusMerged {
		return nil, ErrPRAlreadyMerged
	}

	found := false
	for _, reviewerID := range pr.Reviewers {
		if reviewerID == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrReviewerNotAssigned
	}

	teamName, err := s.teamRepo.GetUserTeam(oldReviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer team: %w", err)
	}
	if teamName == "" {
		return nil, ErrReviewerNotInTeam
	}

	candidates, err := s.userRepo.GetActiveUsersByTeam(teamName, oldReviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	assignedReviewers := make(map[int]struct{})
	for _, reviewerID := range pr.Reviewers {
		if reviewerID != oldReviewerID {
			assignedReviewers[reviewerID] = struct{}{}
		}
	}

	filteredCandidates := make([]models.User, 0)
	for _, candidate := range candidates {
		if candidate.ID != pr.AuthorID {
			if _, alreadyAssigned := assignedReviewers[candidate.ID]; !alreadyAssigned {
				filteredCandidates = append(filteredCandidates, candidate)
			}
		}
	}

	if len(filteredCandidates) == 0 {
		return nil, ErrNoAvailableReviewers
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	newReviewerID := filteredCandidates[r.Intn(len(filteredCandidates))].ID

	if err := s.prRepo.ReassignReviewer(prID, oldReviewerID, newReviewerID); err != nil {
		return nil, fmt.Errorf("failed to reassign reviewer: %w", err)
	}
	updatedPR, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated PR: %w", err)
	}

	return updatedPR, nil
}
