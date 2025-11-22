package service

import (
	"fmt"
	"pr-reviewer-service/internal/repository"
)

type StatsService struct {
	prRepo *repository.PRRepository
}

func NewStatsService(prRepo *repository.PRRepository) *StatsService {
	return &StatsService{prRepo: prRepo}
}

func (s *StatsService) GetStats() (map[string]int, error) {
	stats, err := s.prRepo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}

