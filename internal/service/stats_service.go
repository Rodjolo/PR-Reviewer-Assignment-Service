package service

import (
	"fmt"
	"github.com/Rodjolo/pr-reviewer-service/internal/repository"
	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
)

type StatsService struct {
	prRepo repository.PRRepositoryInterface
}

func NewStatsService(prRepo repository.PRRepositoryInterface) *StatsService {
	return &StatsService{prRepo: prRepo}
}

func (s *StatsService) GetStats() (*dto.StatsResponse, error) {
	statsMap, err := s.prRepo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	
	return &dto.StatsResponse{
		TotalUsers:  statsMap["total_users"],
		ActiveUsers: statsMap["active_users"],
		TotalTeams:  statsMap["total_teams"],
		TotalPRs:    statsMap["total_prs"],
		OpenPRs:     statsMap["open_prs"],
		MergedPRs:   statsMap["merged_prs"],
	}, nil
}
