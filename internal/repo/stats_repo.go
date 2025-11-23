package repo

import "context"

// Статистика
type ReviewerStats struct {
	UserID          string
	AssignmentCount int
}

type StatsRepository interface {
	// GetReviewerStats получает статистику по ревьюверам
	GetReviewerStats(ctx context.Context) ([]ReviewerStats, error)

	// GetPRStats получает статистику по PR
	GetPRStats(ctx context.Context) (map[string]int, error)
}
